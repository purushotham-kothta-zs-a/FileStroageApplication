package model

import "time"

type File struct {
	FileId    string
	FileName  string
	FileSize  int64
	Tags      []string
	CreatedAt time.Time
}

type PatchReq struct {
	FileName *string   `json:"fileName,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}
