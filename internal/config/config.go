package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type Config struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		log.Fatal("not found config path")
	}

	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("cannot red config: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		log.Fatalf("cannot decode config: %s", err)
	}

	if cfg.Floors <= 0 {
		log.Fatal("Floors <= 0")
	}
	if cfg.Monsters <= 0 {
		log.Fatal("Monsters <= 0")
	}
	if cfg.Duration <= 0 {
		log.Fatal("Duration <= 0")
	}

	return &cfg
}
func fetchConfigPath() string {
	var path string

	// --config="path/to/config.json"
	flag.StringVar(&path, "config", "", "path config path")
	flag.Parse()

	return path
}
