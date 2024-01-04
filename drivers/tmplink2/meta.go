package tmplink2

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootID
	// define other
	Token    string `json:"token"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

var config = driver.Config{
	Name:              "Tmplink2",
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
		return &Tmplink2{}
	})
}
