package qingstor

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/google/uuid"
	qconfig "github.com/qingstor/qingstor-sdk-go/v4/config"
	"github.com/qingstor/qingstor-sdk-go/v4/service"
	"github.com/redis/go-redis/v9"
	"io"
	"path"
	"time"
)

type Qingstor struct {
	model.Storage
	Addition
	client  *service.Bucket
	rclient *redis.Client
}

func (d *Qingstor) Config() driver.Config {
	return config
}

func (d *Qingstor) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Qingstor) Init(ctx context.Context) error {
	qc, err := qconfig.New(d.AccessKeyId, d.SecretAccessKey)
	if err != nil {
		return err
	}
	svc, err := service.Init(qc)
	if err != nil {
		return err
	}
	d.client, err = svc.Bucket(d.Bucket, d.Zone)
	if err != nil {
		return err
	}

	d.rclient = redis.NewClient(&redis.Options{
		Addr:     "172.30.1.2:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return nil
}

func (d *Qingstor) Drop(ctx context.Context) error {
	return nil
}

func (d *Qingstor) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	dirId := dir.GetID()
	if dirId == "" {
		dirId = d.Bucket
	}

	members, err := d.rclient.SMembers(ctx, (&QingStorMeta{ParentID: dirId}).getParentKey()).Result()
	if err != nil {
		return nil, err
	}

	var objs []model.Obj
	for _, member := range members {
		var meta QingStorMeta
		metaStr, err := d.rclient.Get(ctx, member).Result()
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(metaStr), &meta)

		objs = append(objs, &model.Object{
			ID:       meta.ID,
			Path:     path.Join(dir.GetPath(), meta.Name),
			Name:     meta.Name,
			Size:     meta.Size,
			Modified: meta.Modified,
			Ctime:    meta.Ctime,
			IsFolder: meta.IsFolder,
		})
	}

	return objs, nil
}

func (d *Qingstor) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var meta QingStorMeta
	metaStr, err := d.rclient.Get(ctx, (&QingStorMeta{ID: file.GetID()}).getKey()).Result()
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(metaStr), &meta)

	req, _, _ := d.client.GetObjectRequest(meta.Path, &service.GetObjectInput{})
	_ = req.Build()
	_ = req.SignQuery(600)
	return &model.Link{
		URL: req.HTTPRequest.URL.String(),
	}, nil
}

func (d *Qingstor) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	parentID := parentDir.GetID()
	if parentID == "" {
		parentID = d.Bucket
	}

	meta := &QingStorMeta{
		ID:       uuid.NewString(),
		Path:     path.Join(parentDir.GetPath(), dirName),
		ParentID: parentID,
		Name:     dirName,
		Size:     0,
		IsFolder: true,
		MD5:      "",
		Modified: time.Now(),
		Ctime:    time.Now(),
	}
	metaStr, _ := json.Marshal(meta)
	if err := d.rclient.Set(ctx, meta.getKey(), metaStr, 0).Err(); err != nil {
		return nil, err
	}
	if err := d.rclient.SAdd(ctx, meta.getParentKey(), meta.getKey()).Err(); err != nil {
		return nil, err
	}

	return &model.Object{
		ID:       meta.ID,
		Path:     path.Join(parentDir.GetPath(), dirName),
		Name:     dirName,
		Size:     0,
		Modified: meta.Modified,
		Ctime:    meta.Ctime,
		IsFolder: true,
		HashInfo: utils.HashInfo{},
	}, nil
}

func (d *Qingstor) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO move obj, optional
	return nil, errs.NotImplement
}

func (d *Qingstor) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	// TODO rename obj, optional
	return nil, errs.NotImplement
}

func (d *Qingstor) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO copy obj, optional
	return nil, errs.NotImplement
}

func (d *Qingstor) Remove(ctx context.Context, obj model.Obj) error {
	meta := &QingStorMeta{
		ID: obj.GetID(),
	}

	if !obj.IsDir() {
		_, err := d.client.DeleteObject(obj.GetPath())
		if err != nil {
			return err
		}
	}

	metaStr, _ := d.rclient.Get(ctx, meta.getKey()).Result()
	_ = json.Unmarshal([]byte(metaStr), &meta)

	if err := d.rclient.SRem(ctx, meta.getParentKey(), meta.getKey()).Err(); err != nil {
		return err
	}

	if err := d.rclient.Del(ctx, meta.getKey()).Err(); err != nil {
		return err
	}

	return nil
}

func (d *Qingstor) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	key := path.Join(dstDir.GetPath(), stream.GetName())
	// const DEFAULT int64 = 10485760
	h := md5.New()
	// need to calculate md5 of the full content
	tempFile, err := stream.CacheFullInTempFile()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tempFile.Close()
	}()
	if _, err = io.Copy(h, tempFile); err != nil {
		return nil, err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	md5Str := hex.EncodeToString(h.Sum(nil))

	contentType := stream.GetMimetype()
	storageClass := "STANDARD"
	contentLen := stream.GetSize()
	_, err = d.client.PutObject(key, &service.PutObjectInput{
		ContentLength:   &contentLen,
		ContentMD5:      &md5Str,      // optional. You can manually calculate this to check uploaded file is intact or not.
		ContentType:     &contentType, // ContentType and ContentLength will be detected automatically if empty
		Body:            stream,
		XQSStorageClass: &storageClass, // optional. default to be “STANDARD”. value can be "STANDARD" or “STANDARD_IA”.
	})
	if err != nil {
		return nil, err
	}

	parentID := dstDir.GetID()
	if parentID == "" {
		parentID = d.Bucket
	}
	meta := &QingStorMeta{
		ID:       uuid.NewString(),
		Path:     path.Join(dstDir.GetPath(), stream.GetName()),
		ParentID: parentID,
		Name:     stream.GetName(),
		IsFolder: false,
		Size:     stream.GetSize(),
		MD5:      md5Str,
		Modified: time.Now(),
		Ctime:    time.Now(),
	}
	metaStr, _ := json.Marshal(meta)
	_ = d.rclient.Set(ctx, meta.getKey(), metaStr, 0).Err()
	_ = d.rclient.SAdd(ctx, meta.getParentKey(), meta.getKey()).Err()

	return &model.Object{
		ID:       meta.ID,
		Path:     path.Join(dstDir.GetPath(), stream.GetName()),
		Name:     stream.GetName(),
		Size:     stream.GetSize(),
		Modified: time.Time{},
		Ctime:    time.Time{},
		IsFolder: false,
	}, nil
}

//func (d *Qingstor) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Qingstor)(nil)
