package handlers

import (
	"TestTaskYadro/internal/model"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func ReadEventsInFile(playInfo *model.PlayInfo) error {
	scanner := bufio.NewScanner(os.Stdin)

	countLine := 0
	for scanner.Scan() {
		countLine++

		line := strings.Split(scanner.Text(), " ")

		if len(line) == 1 && line[0] == "" {
			continue
		}

		incomingEvent, err := parseIncomingEvent(line)
		if err != nil {
			return fmt.Errorf("error parse incoming event: %w, line №%d", err, countLine)
		}

		err = eventHandler(incomingEvent, playInfo)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

	}

	return nil
}

func eventHandler(incomingEvent model.IncomingEvent, playInfo *model.PlayInfo) error {
	switch incomingEvent.EventID {
	case 1:
		if err := playerRegistrationEvent(incomingEvent, playInfo); err != nil {
			return fmt.Errorf("%w", err)
		}
	case 2:
		entranceToTheDungeonEvent(incomingEvent, playInfo)
	case 3:
		playerKilledMonsterEvent(incomingEvent, playInfo)
	case 4:
		playerWentNextFloorEvent(incomingEvent, playInfo)
	case 5:
		playerWentPreviousFloorEvent(incomingEvent, playInfo)
	case 6:
		playerEnteredBossFloorEvent(incomingEvent, playInfo)
	case 7:
		playerKilledBossEvent(incomingEvent, playInfo)
	case 8:
		playerLeftDungeonEvent(incomingEvent, playInfo)
	case 9:
		playerCannotContinue(incomingEvent, playInfo)
	case 10:
		if err := playerRecoveringEvent(incomingEvent, playInfo); err != nil {
			return fmt.Errorf("%w", err)
		}
	case 11:
		if err := playerGetDamageEvent(incomingEvent, playInfo); err != nil {
			return fmt.Errorf("%w", err)
		}
	default:
		return fmt.Errorf("there is no such event")
	}

	return nil
}

func parseIncomingEvent(line []string) (model.IncomingEvent, error) {
	var incomingEvent model.IncomingEvent

	if len(line) < 3 {
		return incomingEvent, fmt.Errorf("error event string")
	}

	timeStr := strings.Trim(line[0], "[]")

	eventTime, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return incomingEvent, fmt.Errorf("error parse time")
	}
	incomingEvent.EventTime = eventTime

	playerID, err := strconv.Atoi(line[1])
	if err != nil {
		return incomingEvent, fmt.Errorf("error parse PlayerID")
	}
	incomingEvent.PlayerID = playerID

	eventID, err := strconv.Atoi(line[2])
	if err != nil {
		return incomingEvent, fmt.Errorf("error parse EventID")
	}
	incomingEvent.EventID = eventID

	if len(line) > 3 {
		incomingEvent.ExtraParam = line[3]
	}

	return incomingEvent, nil
}
