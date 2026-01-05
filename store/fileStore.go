package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krogertechnology/krogo/pkg/krogo"
	"github.com/purushotham-kr/FileHub/model"
)

type FileStore interface {
	Upload(c *krogo.Context, file model.File) error
	GetFileById(c *krogo.Context, fileId string) (*model.File, error)
	DeleteFileById(c *krogo.Context, fileId string) error
	UpdateFileById(c *krogo.Context, fileId string, fileName string, tags []string) error
	GetFileNameById(c *krogo.Context, fileId string) (string, error)
	UpdateFileNameById(c *krogo.Context, fileId string, fileName string) error
	UpdateTagsById(c *krogo.Context, fileId string, tags []string) error
	UpdateFile(c *krogo.Context, file model.File) error
	GetFilesByLimit(c *krogo.Context, limit int, cursor *model.Cursor, fromStart bool) ([]model.File, error)
}

type fileStoreImpl struct{}

func (f fileStoreImpl) GetFilesByLimit(c *krogo.Context, limit int, cursor *model.Cursor, fromStart bool) ([]model.File, error) {
	files := make([]model.File, 0)
	var err error
	var rows *sql.Rows

	if !fromStart {
		rows, err = c.DB().Query("SELECT id,file_name,file_size,file_tags,created_at FROM files WHERE (id ,created_at) > (?,?) ORDER BY id,created_at LIMIT ?", cursor.FileId, cursor.CreatedAt, limit+1)
	} else {
		rows, err = c.DB().Query("SELECT id,file_name,file_size,file_tags,created_at FROM files ORDER BY id,created_at LIMIT ?", limit+1)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		file := model.File{}
		var tags []byte
		err = rows.Scan(&file.FileId, &file.FileName, &file.FileSize, &tags, &file.CreatedAt)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(tags, &file.Tags)

		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (f fileStoreImpl) UpdateFile(c *krogo.Context, file model.File) error {
	jsonTags, err := json.Marshal(file.Tags)

	if err != nil {
		return err
	}
	if _, err := c.DB().Exec("UPDATE files SET  file_tags=?, file_name =? ,file_size=? WHERE id=?", jsonTags, file.FileName, file.FileSize, file.FileId); err != nil {
		return err
	}
	return nil
}

func New() FileStore {
	return &fileStoreImpl{}
}

func (f fileStoreImpl) UpdateFileNameById(c *krogo.Context, fileId string, fileName string) error {
	if _, err := c.DB().Exec("UPDATE files SET file_name =? WHERE id=?", fileName, fileId); err != nil {
		return err
	}
	return nil
}

func (f fileStoreImpl) UpdateTagsById(c *krogo.Context, fileId string, tags []string) error {
	jsonTags, err := json.Marshal(tags)

	if err != nil {
		return err
	}
	if _, err := c.DB().Exec("UPDATE files SET  file_tags=? WHERE id=?", jsonTags, fileId); err != nil {
		return err
	}
	return nil
}

func (f fileStoreImpl) GetFileNameById(c *krogo.Context, fileId string) (string, error) {
	var fileName string
	err := c.DB().QueryRow("SELECT file_name FROM files where id=?", fileId).Scan(&fileName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(fmt.Sprintf("file not found with id %s ", fileId))
		}
		return "", err
	}
	return fileName, nil
}

func (f fileStoreImpl) UpdateFileById(c *krogo.Context, fileId string, fileName string, tags []string) error {
	jsonTags, err := json.Marshal(tags)

	if err != nil {
		return err
	}
	if _, err := c.DB().Exec("UPDATE files SET  file_tags=?, file_name =? WHERE id=?", jsonTags, fileName, fileId); err != nil {
		return err
	}
	return nil
}

func (f fileStoreImpl) Upload(c *krogo.Context, file model.File) error {
	jsonTags, err := json.Marshal(file.Tags)

	if err != nil {
		return err
	}

	_, err = c.DB().Exec("INSERT INTO files(id,file_name,file_size,file_tags) VALUES(?, ?, ?, ?)", file.FileId, file.FileName, file.FileSize, jsonTags)

	if err != nil {
		return err
	}

	return nil
}

func (f fileStoreImpl) GetFileById(c *krogo.Context, fileId string) (*model.File, error) {
	var file model.File
	var jsonTags []byte

	err := c.DB().
		QueryRow("SELECT file_name,file_size,file_tags FROM files WHERE id = ?", fileId).
		Scan(&file.FileName, &file.FileSize, &jsonTags)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &file, errors.New(fmt.Sprintf("file not found with id %s", fileId))
		}
		return nil, err
	}

	err = json.Unmarshal(jsonTags, &file.Tags)

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (f fileStoreImpl) DeleteFileById(c *krogo.Context, fileId string) error {
	_, err := c.DB().Exec("DELETE FROM files WHERE id = ?", fileId)

	if err != nil {
		return err
	}

	return nil
}
