package config

import (
	"encoding/json"
	"flag"
	"fmt"
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

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
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

func (c *Config) Validate() error {
	if c.Floors <= 0 {
		return fmt.Errorf("Floors must be > 0, got %d", c.Floors)
	}
	if c.Monsters <= 0 {
		return fmt.Errorf("Monsters must be > 0, got %d", c.Monsters)
	}
	if c.Duration <= 0 {
		return fmt.Errorf("Duration must be > 0, got %d", c.Duration)
	}
	return nil
}
