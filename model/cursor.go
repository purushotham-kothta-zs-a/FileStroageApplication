package model

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Cursor struct {
	FileId    string    `json:"fileId"`
	CreatedAt time.Time `json:"createdAt"`
}

func EncodeCursor(id string, createdAt time.Time) (string, error) {
	cursor := Cursor{
		FileId:    id,
		CreatedAt: createdAt,
	}

	encodedCursor, err := json.Marshal(cursor)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encodedCursor), nil
}

func DecodeCursor(encodedCursor string) (*Cursor, error) {

	cursorBytes, err := base64.StdEncoding.DecodeString(encodedCursor)
	
	if err != nil {
		return nil, err
	}
	var cursor Cursor

	err = json.Unmarshal(cursorBytes, &cursor)

	if err != nil {
		return nil, err
	}

	return &cursor, nil
}
