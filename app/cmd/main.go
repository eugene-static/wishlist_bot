package main

import (
	"github.com/eugene-static/wishlist_bot/app/internal/server"
	"github.com/eugene-static/wishlist_bot/app/lib/config"
)

func main() {
	cfg, err := config.Get("app/internal/config/config.json")
	if err != nil {
		panic(err)
	}
	server.New(cfg).Start()
}
