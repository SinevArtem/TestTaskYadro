package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"fmt"
	"time"
)

// Incoming events №1
func playerRegistrationEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	if incomingEvent.PlayerID < 1 {
		return fmt.Errorf("not support id")
	}

	playInfo.Players[incomingEvent.PlayerID] = &model.Player{
		ID:         incomingEvent.PlayerID,
		Registered: true,
		Health:     100,
	}

	fmt.Printf("[%s] Player [%d] registered\n",
		incomingEvent.EventTime.Format("15:04:05"),
		incomingEvent.PlayerID)

	return nil
}

// Incoming events №2
func entranceToTheDungeonEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {

	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 2)
		return
	}

	if !isDungeonOpen(incomingEvent.EventTime, playInfo.Config) {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 2)
		return
	}

	player.InDungeon = true
	player.CurrentFloor = 1
	player.CurrentFloorEnterTime = &incomingEvent.EventTime
	player.EnterTime = &incomingEvent.EventTime
	player.KilledOnCurrent = 0

	fmt.Printf("[%s] Player [%d] entered the dungeon\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

}

// Incoming events №3
func playerKilledMonsterEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {

	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 3)
		return
	}

	if player.CurrentFloor == playInfo.Config.Floors+1 {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 3)
		return
	}

	if player.FloorCleared {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 3)
		return
	}

	if player.KilledOnCurrent >= playInfo.Config.Monsters {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 3)
		return
	}

	player.KilledOnCurrent++

	if player.KilledOnCurrent == playInfo.Config.Monsters {
		clearTime := incomingEvent.EventTime.Sub(*player.CurrentFloorEnterTime)
		player.FloorClearTimes = append(player.FloorClearTimes, clearTime)
		player.FloorCleared = true
	}

	fmt.Printf("[%s] Player [%d] killed the monster\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

}

// Incoming events №4
func playerWentNextFloorEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 4)
		return
	}

	if player.CurrentFloor == playInfo.Config.Floors+1 {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 4)
		return
	}

	if !player.FloorCleared {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 4)
		return
	}

	player.CurrentFloor++
	player.CurrentFloorEnterTime = &incomingEvent.EventTime
	player.KilledOnCurrent = 0
	player.FloorCleared = false

	fmt.Printf("[%s] Player [%d] went to the next floor\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

}

// Incoming events №5
func playerWentPreviousFloorEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 5)
		return
	}

	if player.CurrentFloor == 1 {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 5)
		return
	}

	player.CurrentFloor--
	player.CurrentFloorEnterTime = &incomingEvent.EventTime

	fmt.Printf("[%s] Player [%d] went to the previous floor\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)
}

// Incoming events №6
func playerEnteredBossFloorEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 6)
		return
	}

	if player.CurrentFloor == playInfo.Config.Floors+1 {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 6)
		return
	}

	if player.CurrentFloor != playInfo.Config.Floors {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 6)
		return
	}

	if !player.FloorCleared {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 6)
		return
	}

	player.CurrentFloor++
	player.CurrentFloorEnterTime = &incomingEvent.EventTime

	fmt.Printf("[%s] Player [%d] entered the boss's floor\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)
}

//---------------------------------------------------------------------------------------------

// Incoming events №31
func playerDisqualifiedEvent(time time.Time, playerID int) {
	fmt.Printf("[%s] Player [%d] is disqualified\n",
		time.Format("15:04:05"),
		playerID)
}

// Incoming events №32
func playerDeadEvent(time time.Time, playerID int) {
	fmt.Printf("[%s] Player [%d] is dead\n",
		time.Format("15:04:05"),
		playerID)
}

// Incoming events №33
func playerImpossibleMoveEvent(time time.Time, playerID int, eventID int) {
	fmt.Printf("[%s] Player [%d] makes imposible move [%d]\n",
		time.Format("15:04:05"),
		playerID,
		eventID)
}

//---------------------------------------------------------------------------------------------

func isDungeonOpen(eventTime time.Time, cfg *config.Config) bool {
	openTime, _ := time.Parse("15:04:05", cfg.OpenAt)

	openDateTime := time.Date(
		eventTime.Year(), eventTime.Month(), eventTime.Day(),
		openTime.Hour(), openTime.Minute(), openTime.Second(),
		0, eventTime.Location(),
	)

	closeDateTime := openDateTime.Add(time.Duration(cfg.Duration) * time.Hour)

	return (eventTime.Equal(openDateTime) || eventTime.After(openDateTime)) &&
		eventTime.Before(closeDateTime)
}

func validateAndGetPlayer(event model.IncomingEvent, playInfo *model.PlayInfo) *model.Player {
	player, ok := playInfo.Players[event.PlayerID]

	if !ok || !player.Registered {
		playerDisqualifiedEvent(event.EventTime, event.PlayerID)
		return nil
	}

	if player.Dead {
		playerDeadEvent(event.EventTime, event.PlayerID)
		return nil
	}

	if player.Disqualified {
		playerDisqualifiedEvent(event.EventTime, event.PlayerID)
		return nil
	}

	return player
}
