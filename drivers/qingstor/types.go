package qingstor

import "time"

type QingStorMeta struct {
	ID       string    `json:"id"`
	Path     string    `json:"path,omitempty"`
	ParentID string    `json:"parent_id,omitempty"`
	Name     string    `json:"name"`
	Size     int64     `json:"size,omitempty"`
	IsFolder bool      `json:"is_folder,omitempty"`
	MD5      string    `json:"md5,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
	Ctime    time.Time `json:"ctime,omitempty"`
}

func getRootKey(clientID string) string {
	return "qingstor:dir:" + clientID
}

func (m *QingStorMeta) getKey() string {
	return "qingstor:file:" + m.ID
}

func (m *QingStorMeta) getParentKey() string {
	return "qingstor:dir:" + m.ParentID
}
