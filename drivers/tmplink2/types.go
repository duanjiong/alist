package tmplink2

import "github.com/alist-org/alist/v3/internal/model"

var _ model.Obj = (*TmpLinkObj)(nil)

type TmpLinkObj struct {
	Mrid string `json:"mr_id"`
	UKey string `json:"ukey"`
	model.Object
}

type TmpLinkSubMeetingRoomData struct {
	Mrid  string `json:"mr_id"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

type TmpLinkMeetingRoomData struct {
	Mrid     string                      `json:"mr_id"`
	Name     string                      `json:"name"`
	Model    string                      `json:"model"`
	SubRooms []TmpLinkSubMeetingRoomData `json:"sub_rooms"`
}

type TmpLinkMeetingRoomResponse struct {
	Data   TmpLinkMeetingRoomData `json:"data"`
	Status int                    `json:"status"`
	Debug  []string               `json:"debug"`
}

type TmpLinkFileData struct {
	Mrid     string `json:"mrid"`
	Model    int    `json:"model"`
	UKey     string `json:"ukey"`
	SHA1     string `json:"sha1"`
	Fname    string `json:"fname"`
	Fsize    int    `json:"fsize"`
	Ftype    string `json:"ftype"`
	LeftTime int    `json:"lefttime"`
}

type TmpLinkFileResponse struct {
	Data   []TmpLinkFileData `json:"data"`
	Status int               `json:"status"`
	Debug  []string          `json:"debug"`
}
