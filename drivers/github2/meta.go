package github2

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// define other
	Vendor   string `json:"vendor" type:"select" options:"github,gitee" default:"github"`
	Token    string `json:"token"`
	Username string `json:"username" required:"true"`
	Repo     string `json:"repo" required:"true"`
	Public   bool   `json:"public" default:"false"`
	Branch   string `json:"branch" default:"master" required:"true"`
}

var config = driver.Config{
	Name:              "Github2",
	LocalSort:         false,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Github2{}
	})
}
