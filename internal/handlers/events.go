package handlers

import (
	"TestTaskYadro/internal/model"
	"fmt"
	"time"
)

// 1
func playerRegistrationEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	if incomingEvent.PlayerID < 1 {
		return fmt.Errorf("not support id")
	}

	playInfo.Players[incomingEvent.PlayerID] = &model.Player{
		ID:         incomingEvent.PlayerID,
		Registered: true,
		Health:     100,
	}

	return nil
}

// 2
func entranceToTheDungeonEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	if _, ok := playInfo.Players[incomingEvent.PlayerID]; !ok {
		playerDisqualifiedEvent(incomingEvent.EventTime, incomingEvent.PlayerID)
		return nil
	}

	return nil
}

// Outgoing events
// 31
func playerDisqualifiedEvent(time time.Time, playerID int) {
	fmt.Printf("[%s] Player [%d] is disqualified\n",
		time.Format("15:04:05"),
		playerID)
}

// 32
func playerDeadEvent(time time.Time, playerID int) {
	fmt.Printf("[%s] Player [%d] is dead\n",
		time.Format("15:04:05"),
		playerID)
}

// 33
func playerImpossibleMoveEvent(time time.Time, playerID int, eventID int) {
	fmt.Printf("[%s] Player [%d] makes imposible move [%d]\n",
		time.Format("15:04:05"),
		playerID,
		eventID)
}
