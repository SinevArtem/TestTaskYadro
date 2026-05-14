package main

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/handlers"
	"TestTaskYadro/internal/model"
	"log"
)

// go run cmd/app/main.go --config=./config/config.json
func main() {
	cfg := config.MustLoad()

	playInfo := model.PlayInfo{Config: cfg, Players: make(map[int]*model.Player)}

	err := handlers.ReadEventsInFile(&playInfo)
	if err != nil {
		log.Println(err)
	}

	// if err := handlers.FinalReport(&playInfo); err != nil {

	// }

}
