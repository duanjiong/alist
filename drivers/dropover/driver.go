package dropover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"path"
	"strings"
	"time"
)

type Dropover struct {
	model.Storage
	Addition
	client  *http.Client
	rclient *redis.Client
}

func (d *Dropover) Config() driver.Config {
	return config
}

func (d *Dropover) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Dropover) Init(ctx context.Context) error {
	// TODO login / refresh token
	d.client = &http.Client{}
	d.rclient = redis.NewClient(&redis.Options{
		Addr:     "172.30.1.2:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return nil
}

func (d *Dropover) Drop(ctx context.Context) error {
	return nil
}

func (d *Dropover) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	dirId := dir.GetID()
	if dirId == "" {
		dirId = d.ClientID
	}

	members, err := d.rclient.SMembers(ctx, (&DropOverMeta{ParentID: dirId}).getParentKey()).Result()
	if err != nil {
		return nil, err
	}

	var objs []model.Obj
	for _, member := range members {
		var meta DropOverMeta
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

func (d *Dropover) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var meta DropOverMeta
	metaStr, err := d.rclient.Get(ctx, (&DropOverMeta{ID: file.GetID()}).getKey()).Result()
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(metaStr), &meta)
	return &model.Link{
		URL: meta.Link,
	}, nil
}

func (d *Dropover) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	parentID := parentDir.GetID()
	if parentID == "" {
		parentID = d.ClientID
	}

	meta := &DropOverMeta{
		ID:       uuid.NewString(),
		Link:     "",
		ParentID: parentID,
		Name:     dirName,
		Size:     0,
		IsFolder: true,
		SHA256:   "",
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

func (d *Dropover) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO move obj, optional
	return nil, errs.NotImplement
}

func (d *Dropover) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	// TODO rename obj, optional
	return nil, errs.NotImplement
}

func (d *Dropover) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO copy obj, optional
	return nil, errs.NotImplement
}

func (d *Dropover) Remove(ctx context.Context, obj model.Obj) error {
	meta := &DropOverMeta{
		ID: obj.GetID(),
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

func (d *Dropover) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	fullHash := stream.GetHash().GetHash(utils.SHA256)
	if len(fullHash) <= 0 {
		tmpF, err := stream.CacheFullInTempFile()
		if err != nil {
			return nil, err
		}
		fullHash, err = utils.HashFile(utils.SHA256, tmpF)
		if err != nil {
			return nil, err
		}
	}
	fullHash = strings.ToUpper(fullHash)

	data := map[string]interface{}{
		"config": map[string]interface{}{
			"title":     "",
			"link_type": "direct",
		},
		"items": []map[string]interface{}{
			{
				"content_type":   stream.GetMimetype(),
				"file_name":      stream.GetName(),
				"f":              fullHash,
				"content_length": fmt.Sprintf("%d", stream.GetSize()),
			},
		},
	}
	b, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", fmt.Sprintf("https://a.dropoverapp.com/cloud/%s/session", d.ClientID), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "a.dropoverapp.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Dropover/350 CFNetwork/1333.0.4 Darwin/21.5.0")
	req.Header.Set("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&response)
	token := response["token"].(string)
	sessionId := response["id"].(string)
	fileId := response["items"].([]interface{})[0].(map[string]interface{})["id"].(string)

	// step 2: get upload url
	req, _ = http.NewRequest("PUT", fmt.Sprintf("https://a.dropoverapp.com/cloud/%s/session/%s/items/%s/upload_url", d.ClientID, sessionId, fileId), nil)
	req.Header.Set("Host", "a.dropoverapp.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Dropover/350 CFNetwork/1333.0.4 Darwin/21.5.0")
	req.Header.Set("X-DROPOVER-CLOUD-SCP", "1")
	req.Header.Set("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Set("X-DROPOVER-SESSION-TOKEN", token)
	resp, _ = d.client.Do(req)
	_ = json.NewDecoder(resp.Body).Decode(&response)
	uploadUrl := response["url"].(string)

	// step 3: upload
	req, _ = http.NewRequest("PUT", uploadUrl, stream)
	req.Header.Set("Content-Type", stream.GetMimetype())
	req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("x-amz-acl", "public-read")
	//req.Header.Set("x-amz-content-sha256", fullHash)
	req.Header.Set("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", stream.GetName()))
	req.Header.Set("User-Agent", "Dropover/350 CFNetwork/1333.0.4 Darwin/21.5.0")
	req.TransferEncoding = []string{"identity"}
	req.ContentLength = stream.GetSize()
	resp, _ = d.client.Do(req)
	_ = json.NewDecoder(resp.Body).Decode(&response)
	etag := resp.Header.Get("ETag")
	if etag == "" {
		return nil, errs.NewErr(err, "etag is empty")
	}

	// step4 : notify server
	data = map[string]interface{}{"etag": etag}
	b, _ = json.Marshal(data)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("https://a.dropoverapp.com/cloud/%s/session/%s/items/%s/finalize", d.ClientID, sessionId, fileId), bytes.NewReader(b))
	req.Header.Set("Host", "a.dropoverapp.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Dropover/350 CFNetwork/1333.0.4 Darwin/21.5.0")
	req.Header.Set("X-DROPOVER-CLOUD-SCP", "1")
	req.Header.Set("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Set("X-DROPOVER-SESSION-TOKEN", token)
	resp, _ = d.client.Do(req)
	_ = json.NewDecoder(resp.Body).Decode(&response)

	linkUrl := response["url"].(string)
	if linkUrl == "" {
		return nil, errs.NewErr(err, "link url is empty")
	}

	// step5: write redis file metadata
	parentID := dstDir.GetID()
	if parentID == "" {
		parentID = d.ClientID
	}
	meta := &DropOverMeta{
		ID:       fileId,
		Link:     linkUrl,
		ParentID: parentID,
		Name:     stream.GetName(),
		IsFolder: false,
		Size:     stream.GetSize(),
		SHA256:   fullHash,
		Modified: time.Now(),
		Ctime:    time.Now(),
	}
	metaStr, _ := json.Marshal(meta)
	_ = d.rclient.Set(ctx, meta.getKey(), metaStr, 0).Err()
	_ = d.rclient.SAdd(ctx, meta.getParentKey(), meta.getKey()).Err()

	// step6: write redis dir metadata

	return &model.Object{
		ID:       fileId,
		Path:     path.Join(dstDir.GetPath(), stream.GetName()),
		Name:     stream.GetName(),
		Size:     stream.GetSize(),
		Modified: time.Time{},
		Ctime:    time.Time{},
		IsFolder: false,
	}, nil
}

//func (d *Dropover) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Dropover)(nil)
