package qingstor

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	AccessKeyId     string `json:"access_key_id" required:"true"`
	SecretAccessKey string `json:"secret_access_key" required:"true"`
	Bucket          string `json:"bucket" required:"true"`
	Zone            string `json:"zone" required:"true"`
}

var config = driver.Config{
	Name:              "Qingstor",
	LocalSort:         false,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Qingstor{}
	})
}
