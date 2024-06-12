package main

import (
	"github.com/eugene-static/wishlist_bot/internal/server"
	"github.com/eugene-static/wishlist_bot/lib/config"
)

func main() {
	cfg, err := config.Get("app/internal/config/config.json")
	if err != nil {
		panic(err)
	}
	server.New(cfg).Start()
}
