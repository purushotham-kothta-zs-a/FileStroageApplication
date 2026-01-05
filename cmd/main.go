package main

import (
	"github.com/krogertechnology/krogo/pkg/krogo"
	"github.com/purushotham-kr/FileHub/cmd/configs"
	"github.com/purushotham-kr/FileHub/handler"
	"github.com/purushotham-kr/FileHub/store"
)

func main() {
	k := krogo.NewCMD()

	config := configs.LoadConfig()
	s := store.New()
	handler := handler.New(s, config)

	k.GET("subscribe", handler.Subscribe)
	k.Start()
}
