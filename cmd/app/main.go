package main

import (
	"TestTaskYadro/internal/config"
	"fmt"
)

// go run cmd/app/main.go --config=./config/config.json
func main() {
	cfg := config.MustLoad()

	fmt.Println(*cfg)
}
