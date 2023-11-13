package main

import (
	"github.com/AhegaoHD/WBT/config"
	"github.com/AhegaoHD/WBT/internal/bootstrap"
	"log"
	"os"
)

func main() {
	configPath := findConfigPath()

	cfg, err := config.Parse(configPath)
	if err != nil {
		log.Fatal(err)
	}

	bootstrap.Run(cfg)

}

func findConfigPath() string {
	const (
		devConfig  = "config/dev.config.toml"
		prodConfig = "config/config.toml"
	)

	if os.Getenv("CFG") == "DEV" {
		return devConfig
	}

	return prodConfig
}
