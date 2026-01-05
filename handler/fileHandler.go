package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/krogertechnology/krogo/pkg/krogo"
	"github.com/krogertechnology/krogo/pkg/krogo/template"
	"github.com/purushotham-kr/FileHub/cmd/configs"
	"github.com/purushotham-kr/FileHub/model"
	"github.com/purushotham-kr/FileHub/store"
)

type FileHandler struct {
	config *configs.Config
	fs     store.FileStore
}

func New(fs store.FileStore, config *configs.Config) *FileHandler {
	return &FileHandler{
		fs:     fs,
		config: config,
	}
}

func (f FileHandler) DownloadFileById(c *krogo.Context) (interface{}, error) {
	fileId := c.PathParam("id")
	fileName, err := f.fs.GetFileNameById(c, fileId)

	if err != nil {
		return nil, err
	}

	isExists, err := f.isFileExistsByName(fileName)

	if err != nil {
		return nil, err
	}

	if !isExists {
		return nil, fmt.Errorf("file %s does not exist", fileName)
	}

	data, err := os.ReadFile(path.Join(f.config.FileStoragePath, fileName))

	if err != nil {
		return nil, err
	}

	c.SetResponseHeader(map[string]string{
		"Content-Disposition": "attachment;" + " filename = " + fileName,
	})

	return template.File{
		Content:     data,
		ContentType: "application/octet-stream",
	}, nil
}

func (f FileHandler) GetFileById(c *krogo.Context) (interface{}, error) {
	fileId := c.PathParam("id")

	file, err := f.fs.GetFileById(c, fileId)

	if err != nil {
		return nil, err
	}

	return *file, err
}

func (f FileHandler) Subscribe(c *krogo.Context) (interface{}, error) {
	for {
		msg, err := c.Subscribe(nil)

		//msg, err := c.SubscribeWithCommit(func(message *pubsub.Message) (bool, bool) {
		//	return true, false
		//})
		fmt.Println("Partition: ", msg.Partition, " ", "offset: ", msg.Offset)
		if err != nil {
			fmt.Println("here")
			fmt.Println(err)
			continue
		}

		operation := msg.Headers["Operation"]

		switch operation {
		case "DELETE":
			var file model.File

			err := json.Unmarshal([]byte(msg.Value), &file)

			if err != nil {
				fmt.Println(err)
			}

			fmt.Println("Operation: ", operation, "FileName: ", file.FileName, "partition: ", msg.Partition, "offset: ", msg.Offset)

			err = f.fs.DeleteFileById(c, file.FileId)

			if err != nil {
				continue
			}
		case "POST":
			var file model.File
			fmt.Println("Operation: ", operation, "FileName: ", msg.Value, "partition: ", msg.Partition, "offset: ", msg.Offset)

			err = json.Unmarshal([]byte(msg.Value), &file)
			if err != nil {
				continue
			}

			err := f.fs.Upload(c, file)
			if err != nil {
				continue
			}
		case "PUT":
			fmt.Println("Operation: ", operation, "FileName: ", msg.Value, "partition: ", msg.Partition, "offset: ", msg.Offset)

			var file model.File

			err = json.Unmarshal([]byte(msg.Value), &file)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err := f.fs.UpdateFile(c, file)
			if err != nil {
				continue
			}

		case "PATCH":
			fmt.Println("Operation: ", operation, "partition: ", msg.Partition, "offset: ", msg.Offset)

			var patchRequestBody model.PatchReq

			err = json.Unmarshal([]byte(msg.Value), &patchRequestBody)

			if err != nil {
				fmt.Println(err)
				continue
			}

			err := f.UpdateFileById(c, msg.Headers["id"], patchRequestBody)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func (f FileHandler) UpdateFileById(c *krogo.Context, fileId string, patchRequestBody model.PatchReq) error {

	isFileNameUpdated := patchRequestBody.FileName != nil
	isTagsUpdated := patchRequestBody.Tags != nil

	var err error

	if isFileNameUpdated && isTagsUpdated {
		err = f.fs.UpdateFileById(c, fileId, *patchRequestBody.FileName, *patchRequestBody.Tags)
	} else if isFileNameUpdated {
		err = f.fs.UpdateFileNameById(c, fileId, *patchRequestBody.FileName)
	} else if isTagsUpdated {
		err = f.fs.UpdateTagsById(c, fileId, *patchRequestBody.Tags)
	}

	if err != nil {
		return err
	}

	return nil
}

func (f FileHandler) AcceptFile(c *krogo.Context) (interface{}, error) {
	err := c.Request().ParseMultipartForm(10 << 20)

	if err != nil {
		return nil, err
	}

	file, fileHeader, err := c.Request().FormFile("file")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	if val, err := f.isFileExistsByName(fileHeader.Filename); err != nil {
		return nil, fmt.Errorf("something went wrong while uploading file")
	} else if val == true {
		return nil, fmt.Errorf("file %s already exists", fileHeader.Filename)
	}

	if err := f.createOrUpdateFile(c, fileHeader.Filename); err != nil {
		return nil, fmt.Errorf("something went wrong while uploading file")
	}

	tags := c.Request().MultipartForm.Value["tags"]

	if tags == nil {
		tags = []string{}
	}

	fileId := uuid.New().String()

	var fileToPublish model.File = model.File{
		FileId:   fileId,
		FileName: fileHeader.Filename,
		FileSize: fileHeader.Size,
		Tags:     tags,
	}

	err = c.PublishEvent(fileId, fileToPublish, map[string]string{
		"id":        fileId,
		"Operation": "POST",
	})

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"status code": http.StatusCreated,
		"id":          fileId,
	}, nil
}

func (f FileHandler) createOrUpdateFile(c *krogo.Context, fileName string) error {
	file, fileHeader, err := c.Request().FormFile("file")

	if err != nil {
		return err
	}
	defer file.Close()

	fp, err := os.Create(path.Join(f.config.FileStoragePath, fileName))

	if err != nil {
		return err
	}

	_, err = io.Copy(fp, file)

	if err != nil {
		return err
	}

	err = os.Rename(path.Join(f.config.FileStoragePath, fileName), path.Join(f.config.FileStoragePath, fileHeader.Filename))

	if err != nil {
		return err
	}

	return nil
}

func (f FileHandler) HandlePatch(c *krogo.Context) (interface{}, error) {
	fileId := c.PathParam("id")

	var patchRequestBody model.PatchReq
	err := c.Bind(&patchRequestBody)

	if err != nil {
		return nil, err
	}

	var fileName string

	if patchRequestBody.FileName != nil {
		updatableName := *patchRequestBody.FileName

		if len(updatableName) == 0 {
			return nil, fmt.Errorf("file name should not be empty")
		}

		fileName, err = f.fs.GetFileNameById(c, fileId)

		if err != nil {
			return nil, err
		}

		if updatableName == fileName {
			patchRequestBody.FileName = nil
		} else {
			isFileExists, err := f.isFileExistsByName(updatableName)

			if err != nil {
				return nil, err
			}

			if isFileExists {
				return nil, fmt.Errorf("file %s already exists", updatableName)
			}

			err = os.Rename(path.Join(f.config.FileStoragePath, fileName), path.Join(f.config.FileStoragePath, *patchRequestBody.FileName))
			if err != nil {
				return nil, err
			}
		}
	}

	if patchRequestBody.FileName != nil || patchRequestBody.Tags != nil {

		err = c.PublishEvent(fileId, patchRequestBody, map[string]string{
			"id":        fileId,
			"Operation": "PATCH",
		})

		if err != nil {
			return nil, err
		}
	}
	return http.StatusCreated, nil
}

func (f FileHandler) DeleteFileById(c *krogo.Context) (interface{}, error) {
	fileId := c.PathParam("id")

	fileName, err := f.fs.GetFileNameById(c, fileId)

	if err != nil {
		return nil, err
	}

	if err := f.deleteFileByName(fileName); err != nil {
		return nil, err
	}

	fileToDelete := model.File{
		FileId: fileId,
	}

	if err := c.PublishEvent(fileId, fileToDelete, map[string]string{
		"id":        fileId,
		"Operation": "DELETE",
	}); err != nil {
		return nil, err
	}

	return http.StatusNoContent, nil
}

func (f FileHandler) deleteFileByName(fileName string) error {
	if err := os.Remove(path.Join(f.config.FileStoragePath, fileName)); err != nil {
		return err
	}
	return nil
}

func (f FileHandler) isFileExistsByName(fileName string) (bool, error) {

	dp, err := os.Open(f.config.FileStoragePath)

	if err != nil {
		return false, err
	}

	defer dp.Close()

	fileInfoList, err := dp.Readdir(-1)

	if err != nil {
		return false, err
	}

	for _, fileInfo := range fileInfoList {
		if fileInfo.Name() == fileName {
			return true, nil
		}
	}

	return false, nil
}

func (f FileHandler) UpdateFile(c *krogo.Context) (interface{}, error) {
	fileId := c.PathParam("id")
	err := c.Request().ParseMultipartForm(10 << 20)

	if err != nil {
		return nil, err
	}

	file, fileHeader, err := c.Request().FormFile("file")

	if err != nil {
		return nil, err
	}

	defer file.Close()
	fileName, err := f.fs.GetFileNameById(c, fileId)

	if err != nil {
		return nil, err
	}

	if val, err := f.isFileExistsByName(fileHeader.Filename); err != nil {
		return nil, err
	} else if val == true {
		return nil, fmt.Errorf("file with name %s already exists", fileHeader.Filename)
	}

	if err := f.createOrUpdateFile(c, fileName); err != nil {
		return nil, err
	}

	tags := c.Request().MultipartForm.Value["tags"]

	if tags == nil {
		tags = []string{}
	}

	var fileToPublish model.File = model.File{
		FileId:   fileId,
		FileName: fileHeader.Filename,
		FileSize: fileHeader.Size,
		Tags:     tags,
	}

	err = c.PublishEvent(fileId, fileToPublish, map[string]string{
		"id":        fileId,
		"Operation": "PUT",
	})

	if err != nil {
		return nil, err
	}

	return http.StatusCreated, nil
}

func (f FileHandler) GetPaginatedFiles(c *krogo.Context) (interface{}, error) {
	params := c.Params()

	limit, _ := strconv.Atoi(params["limit"])

	if limit == 0 {
		limit = 10
	}

	cursor := params["cursor"]

	var files []model.File
	var err error
	cur := &model.Cursor{}

	if cursor != "" {
		cur, err = model.DecodeCursor(params["cursor"])
		if err != nil {
			return nil, err
		}
		files, err = f.fs.GetFilesByLimit(c, limit, cur, false)

	} else {
		files, err = f.fs.GetFilesByLimit(c, limit, cur, true)
	}

	if err != nil {
		return nil, err
	}

	hasMore := len(files) > limit
	if hasMore {
		files = files[:limit]
	}

	if hasMore {
		last := files[len(files)-1]
		nextCursor, _ := model.EncodeCursor(last.FileId, last.CreatedAt)

		return map[string]interface{}{
			"data":       files[:limit],
			"hasMore":    true,
			"nextCursor": nextCursor,
		}, nil
	}

	return map[string]interface{}{
		"data":       files,
		"hasMore":    false,
		"nextCursor": "",
	}, nil
}
