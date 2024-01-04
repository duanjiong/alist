package dropover

import "time"

type DropOverMeta struct {
	ID       string    `json:"id"`
	Link     string    `json:"link,omitempty"`
	ParentID string    `json:"parent_id,omitempty"`
	Name     string    `json:"name"`
	Size     int64     `json:"size,omitempty"`
	IsFolder bool      `json:"is_folder,omitempty"`
	SHA256   string    `json:"sha256,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
	Ctime    time.Time `json:"ctime,omitempty"`
}

func getRootKey(clientID string) string {
	return "dropover:dir:" + clientID
}

func (m *DropOverMeta) getKey() string {
	return "dropover:file:" + m.ID
}

func (m *DropOverMeta) getParentKey() string {
	return "dropover:dir:" + m.ParentID
}
