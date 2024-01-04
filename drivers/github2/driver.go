package github2

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"path"

	"github.com/google/go-github/v57/github"
)

type Github2 struct {
	model.Storage
	Addition
	client *github.Client
}

func (d *Github2) Config() driver.Config {
	return config
}

func (d *Github2) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Github2) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	d.client = github.NewClient(nil).WithAuthToken(d.Addition.Token)
	return nil
}

func (d *Github2) Drop(ctx context.Context) error {
	return nil
}

func (d *Github2) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var result []model.Obj
	_, files, _, err := d.client.Repositories.GetContents(ctx, d.Addition.Username, d.Addition.Repo, dir.GetPath(), &github.RepositoryContentGetOptions{
		Ref: d.Branch,
	})
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		result = append(result, &model.Object{
			Name:     *file.Name,
			Size:     int64(*file.Size),
			IsFolder: *file.Type == "dir",
			HashInfo: utils.NewHashInfo(utils.SHA256, *file.SHA),
		})
	}

	return result, nil
}

func (d *Github2) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var url = ""
	if d.Public {
		url = fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s", d.Addition.Username, d.Addition.Repo, d.Branch, file.GetPath()[1:])
	} else {
		f, _, _, err := d.client.Repositories.GetContents(ctx, d.Addition.Username, d.Addition.Repo, file.GetPath(), &github.RepositoryContentGetOptions{
			Ref: d.Branch,
		})
		if err != nil {
			return nil, err
		}
		url = *f.DownloadURL
	}
	return &model.Link{
		URL: url,
	}, nil
}

func (d *Github2) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	message := "create folder"
	_, _, err := d.client.Repositories.CreateFile(ctx, d.Addition.Username, d.Addition.Repo, path.Join(parentDir.GetPath(), dirName, "init.md")[1:], &github.RepositoryContentFileOptions{
		Content: []byte("init"),
		Branch:  &d.Branch,
		Message: &message,
	})
	if err != nil {
		return nil, err
	}
	return &model.Object{
		Name:     dirName,
		Path:     path.Join(parentDir.GetPath(), dirName),
		IsFolder: true,
	}, nil
}

func (d *Github2) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO move obj, optional
	return nil, errs.NotImplement
}

func (d *Github2) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	// TODO rename obj, optional
	return nil, errs.NotImplement
}

func (d *Github2) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO copy obj, optional
	return nil, errs.NotImplement
}

func (d *Github2) Remove(ctx context.Context, obj model.Obj) error {
	if obj.IsDir() {
		// TODO remove folder, optional
		// github 目前是目录下文件都删掉了就自动删掉了
		return errs.NotImplement
	}
	message := "delete file"
	sha := obj.GetHash().GetHash(utils.SHA256)
	_, _, err := d.client.Repositories.DeleteFile(ctx, d.Addition.Username, d.Addition.Repo, obj.GetPath()[1:], &github.RepositoryContentFileOptions{
		Message: &message,
		Branch:  &d.Branch,
		SHA:     &sha,
	})
	return err
}

func (d *Github2) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	//read from stream and write all date in date bytes
	var toWrite []byte
	tmp := make([]byte, 2048)
	n, err := stream.Read(tmp)
	for err == nil {
		toWrite = append(toWrite, tmp[:n]...)
		n, err = stream.Read(tmp)
	}

	message := "create/or file"
	content, _, err := d.client.Repositories.UpdateFile(ctx, d.Addition.Username, d.Addition.Repo, path.Join(dstDir.GetPath(), stream.GetName())[1:], &github.RepositoryContentFileOptions{
		Message: &message,
		Branch:  &d.Branch,
		Content: toWrite,
	})
	if err != nil {
		return nil, err
	}
	return &model.Object{
		Name:     stream.GetName(),
		Path:     path.Join(dstDir.GetPath(), stream.GetName()),
		Size:     int64(len(toWrite)),
		HashInfo: utils.NewHashInfo(utils.SHA256, *content.Content.SHA),
	}, nil
}

//func (d *Github2) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Github2)(nil)
