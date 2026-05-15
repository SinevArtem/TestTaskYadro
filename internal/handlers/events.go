package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"fmt"
	"strconv"
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
		Status:     0,
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

	if player.CurrentFloor == playInfo.Config.Floors {
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

	if player.CurrentFloor >= playInfo.Config.Floors {
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
	player.FloorCleared = true

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

	if player.CurrentFloor == playInfo.Config.Floors {
		player.CurrentFloorEnterTime = &incomingEvent.EventTime
		player.BossFloorEnterTime = &incomingEvent.EventTime
		fmt.Printf("[%s] Player [%d] entered the boss's floor\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

	} else {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 6)
		return
	}

}

// Incoming events №7
func playerKilledBossEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 7)
		return
	}

	if player.CurrentFloor != playInfo.Config.Floors {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 7)
		return
	}

	if player.BossDefeated {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 7)
		return
	}

	player.BossDefeated = true
	player.BossKillTime = &incomingEvent.EventTime
	player.Status = model.SUCCESS

	fmt.Printf("[%s] Player [%d] killed the boss\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

}

// Incoming events №8
func playerLeftDungeonEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 8)
		return
	}

	player.InDungeon = false
	player.LeaveTime = &incomingEvent.EventTime

	fmt.Printf("[%s] Player [%d] left the dungeon\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID)

}

// Incoming events №9
func playerCannotContinue(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) {

	player, ok := playInfo.Players[incomingEvent.PlayerID]

	if !ok || !player.Registered {
		playerDisqualifiedEvent(incomingEvent.EventTime, incomingEvent.PlayerID, playInfo)
		return
	}

	if player.Status == model.DISQUAL || player.Status == model.FAIL {
		return
	}

	player.Status = model.DISQUAL
	player.LeaveTime = &incomingEvent.EventTime

	playerDisqualifiedEvent(incomingEvent.EventTime, incomingEvent.PlayerID, playInfo)
}

// Incoming events №10
func playerRecoveringEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return nil
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 10)
		return nil
	}

	health, err := strconv.Atoi(incomingEvent.ExtraParam)
	if err != nil {
		return fmt.Errorf("error parse health count")
	}

	player.Health += health
	if player.Health > 100 {
		player.Health = 100
	}

	fmt.Printf("[%s] Player [%d] has restored [%d] of health\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID, health)

	return nil

}

// Incoming events №11
func playerGetDamageEvent(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	player := validateAndGetPlayer(incomingEvent, playInfo)
	if player == nil {
		return nil
	}

	if !player.InDungeon {
		playerImpossibleMoveEvent(incomingEvent.EventTime, incomingEvent.PlayerID, 11)
		return nil
	}

	damage, err := strconv.Atoi(incomingEvent.ExtraParam)
	if err != nil {
		return fmt.Errorf("error parse damage count")
	}

	player.Health -= damage

	fmt.Printf("[%s] Player [%d] recieved [%d] of damage\n", incomingEvent.EventTime.Format("15:04:05"), incomingEvent.PlayerID, damage)

	if player.Health <= 0 {
		player.Health = 0
		player.Status = model.FAIL
		player.LeaveTime = &incomingEvent.EventTime
		playerDeadEvent(incomingEvent.EventTime, incomingEvent.PlayerID)
	}

	return nil
}

//=============================================================================================

// Outgoing events №31
func playerDisqualifiedEvent(time time.Time, playerID int, playInfo *model.PlayInfo) {

	if _, ok := playInfo.Players[playerID]; !ok {
		playInfo.Players[playerID] = &model.Player{
			ID:         playerID,
			Registered: false,
			Status:     model.DISQUAL,
			Health:     100,
		}
	} else {
		playInfo.Players[playerID].Status = model.DISQUAL
	}

	fmt.Printf("[%s] Player [%d] is disqualified\n",
		time.Format("15:04:05"),
		playerID)
}

// Outgoing events №32
func playerDeadEvent(time time.Time, playerID int) {
	fmt.Printf("[%s] Player [%d] is dead\n",
		time.Format("15:04:05"),
		playerID)
}

// Outgoing events №33
func playerImpossibleMoveEvent(time time.Time, playerID int, eventID int) {
	fmt.Printf("[%s] Player [%d] makes imposible move [%d]\n",
		time.Format("15:04:05"),
		playerID,
		eventID)
}

//=============================================================================================

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

	if !ok {
		playerDisqualifiedEvent(event.EventTime, event.PlayerID, playInfo)
		return nil
	}

	if player.Status == model.DISQUAL {
		return nil
	}

	if !player.Registered {
		playerDisqualifiedEvent(event.EventTime, event.PlayerID, playInfo)
		return nil
	}

	if player.Status == model.FAIL {
		playerDeadEvent(event.EventTime, event.PlayerID)
		return nil
	}

	return player
}
