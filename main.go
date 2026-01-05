package main

import (
	"github.com/krogertechnology/krogo/pkg/krogo"
	"github.com/purushotham-kr/FileHub/cmd/configs"
	"github.com/purushotham-kr/FileHub/handler"
	"github.com/purushotham-kr/FileHub/store"
)

func main() {
	k := krogo.New()
	k.Server.ValidateHeaders = false

	config := configs.LoadConfig()

	s := store.New()
	h := handler.New(s, config)

	k.POST("/files", h.AcceptFile)
	k.GET("/files/{id}", h.GetFileById)
	k.GET("/files/{id}/content", h.DownloadFileById)
	k.DELETE("/files/{id}", h.DeleteFileById)
	k.PUT("/files/{id}", h.UpdateFile)
	k.PATCH("/files/{id}", h.HandlePatch)
	k.GET("/files", h.GetPaginatedFiles)
	k.Start()
}
