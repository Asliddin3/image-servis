package main

import (
	"github.com/Asliddin3/image-servis/config"
	"github.com/Asliddin3/image-servis/internal/app"
)

func main() {
	cfg := config.LoadConfig()
	app.Run(cfg)
}
