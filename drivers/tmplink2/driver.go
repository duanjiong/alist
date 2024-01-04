package tmplink2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Tmplink2 struct {
	model.Storage
	Addition
	client        *http.Client
	downloadToken string
}

func (d *Tmplink2) Config() driver.Config {
	return config
}

func (d *Tmplink2) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Tmplink2) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	d.client = &http.Client{}
	var err error
	d.Token, err = d.tokenToken()
	if err != nil {
		return err
	}

	go func() {
		// refresh token every 1 hours
		for {
			time.Sleep(1 * time.Hour)
			d.login()
		}
	}()

	return d.login()
}

func (d *Tmplink2) Drop(ctx context.Context) error {
	return nil
}

func (d *Tmplink2) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var result []TmpLinkObj

	mrs, err := d.listRooms(dir.GetID())
	if err != nil {
		return nil, err
	}
	for _, mr := range mrs {
		result = append(result, TmpLinkObj{
			Mrid: mr.Mrid,
			Object: model.Object{
				Name:     mr.Name,
				ID:       mr.Mrid,
				IsFolder: true,
			},
		})
	}
	files, err := d.listFilesInRoom(dir.GetID())
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		result = append(result, TmpLinkObj{
			Mrid: file.Mrid,
			UKey: file.UKey,
			Object: model.Object{
				Name:     file.Fname,
				ID:       file.UKey,
				Size:     int64(file.Fsize),
				IsFolder: false,
			},
		})
	}

	return utils.SliceConvert(result, func(src TmpLinkObj) (model.Obj, error) {
		return &src, nil
	})
}

func (d *Tmplink2) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	captcha, err := d.tokenChallenge()
	if err != nil {
		return nil, err
	}
	token, err := d.tokenToken()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("action", "download_req")
	params.Set("captcha", captcha)
	params.Set("ukey", file.GetID())
	params.Set("token", token)
	body := bytes.NewBufferString(params.Encode())

	// Create a POST request with the buffer content
	request, err := http.NewRequest("POST", "https://tmp-api.vx-cdn.com/api_v2/file", body)
	if err != nil {
		return nil, err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var tokenResponse map[string]interface{}
	_ = json.NewDecoder(response.Body).Decode(&tokenResponse)

	return &model.Link{
		URL: tokenResponse["data"].(string),
	}, nil
}

func (d *Tmplink2) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	// TODO create folder, optional
	return nil, errs.NotImplement
}

func (d *Tmplink2) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO move obj, optional
	return nil, errs.NotImplement
}

func (d *Tmplink2) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	// TODO rename obj, optional
	return nil, errs.NotImplement
}

func (d *Tmplink2) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO copy obj, optional
	return nil, errs.NotImplement
}

func (d *Tmplink2) Remove(ctx context.Context, obj model.Obj) error {
	payload := "action=remove_from_workspace&token=" + d.Token + "&ukey=" + obj.GetID()
	url := "https://tmp-api.vx-cdn.com/api_v2/file"
	if obj.IsDir() {
		payload = "action=delete&token=" + d.Token + "&mr_id=" + obj.GetID()
		url = "https://tmp-api.vx-cdn.com/api_v2/meetingroom"
	}

	request, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return err
	}

	setHeader(request, map[string]string{
		"Host":               "tmp-api.vx-cdn.com",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Not_A Brand";v="8", "Chromium";v="120"`,
		"Accept":             "*/*",
		"DNT":                "1",
		"sec-ch-ua-mobile":   "?0",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"sec-ch-ua-platform": `"macOS"`,
		"Origin":             "https://www.tmp.link",
		"Sec-Fetch-Site":     "cross-site",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://www.tmp.link/",
		"Accept-Language":    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
	})

	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

// "https://tmp-cli.vx-cdn.com/app/upload_cli"
func (d *Tmplink2) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	url := "https://tmp-cli.vx-cdn.com/app/upload_cli"

	// Create a buffer to store the file content
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create form fields
	_ = writer.WriteField("token", d.Token)
	_ = writer.WriteField("model", "2")

	// Create a form file field for the file
	fileWriter, err := writer.CreateFormFile("file", stream.GetName())
	if err != nil {
		return nil, err
	}

	// Copy the file content to the form file field
	_, err = io.Copy(fileWriter, stream)
	if err != nil {
		return nil, err
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Create a POST request with the buffer content
	request, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return nil, err
	}

	// Set the Content-Type header for the request
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Set additional headers
	request.Header.Set("Host", "tmp-cli.vx-cdn.com")
	request.Header.Set("User-Agent", "curl/7.79.1")
	request.Header.Set("Accept", "*/*")

	// Perform the request
	response, err := d.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//response body 内容是一下内容，从中提取 Download Page url
	//	Welcome to use tmp.link upload service.
	//Upload Complete.
	//Download Page: https://ttttt.link/f/659b6c615ba0c
	//Thank you for using our services.
	reader := bufio.NewReader(response.Body)
	for {
		line, err := reader.ReadBytes('\n') // Change '\n' to '\r\n' if you're working with Windows-style line endings
		if err != nil {
			break // Break the loop on EOF or any other error
		}
		if bytes.Contains(line, []byte("Download Page:")) {
			match := "https://ttttt.link/f/"
			begin := bytes.Index(line, []byte(match)) + len(match)
			ukey := strings.TrimSpace(string(line[begin:]))

			err = d.moveFileToMeetingRoom(ukey, dstDir.GetID())
			if err != nil {
				return nil, err
			}

			return &TmpLinkObj{
				Mrid: "",
				UKey: ukey,
				Object: model.Object{
					ID:       ukey,
					Path:     path.Join(dstDir.GetPath(), stream.GetName()),
					Name:     stream.GetName(),
					Size:     stream.GetSize(),
					Modified: time.Time{},
					Ctime:    time.Time{},
					IsFolder: false,
					HashInfo: utils.HashInfo{},
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("not found download page url")

}

//func (d *Tmplink2) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Tmplink2)(nil)
