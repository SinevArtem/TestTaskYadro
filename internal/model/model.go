package model

import (
	"TestTaskYadro/internal/config"
	"time"
)

type PlayInfo struct {
	Players map[int]*Player
	Config  *config.Config
}

type Player struct {
	ID                    int
	Registered            bool
	InDungeon             bool            // в подземелье
	CurrentFloor          int             // 1-based, этаж босса = Floors+1
	KilledOnCurrent       int             // сколько монстров убито на текущем этаже
	Health                int             // 0-100
	EnterTime             *time.Time      // когда вошёл в подземелье
	CurrentFloorEnterTime *time.Time      // когда вошёл на текущий этаж
	FloorCleared          bool            // зачищен ли текущий этаж
	FloorClearTimes       []time.Duration // время зачистки каждого обычного этажа
	BossDefeated          bool            // убит ли босс
	BossKillTime          *time.Time      // когда убил босса
	LeaveTime             *time.Time      // когда покинул
	BossFloorEnterTime    *time.Time      // когда вошёл на этаж босса
	Disqualified          bool
	Dead                  bool
	Status                Status // SUCCESS, FAIL, DISQUAL
}

type IncomingEvent struct {
	EventTime  time.Time
	PlayerID   int
	EventID    int
	ExtraParam string
}

type PlayerReport struct {
	State             string
	PlayerID          int
	TimeInTheDungeon  time.Duration // время проведенное в подземелье
	AverageFloorTime  time.Duration // среднее время этажа
	TimeToKillTheBoss time.Duration // вермя на босса
	Health            int
}

type Status int

const (
	SUCCESS Status = iota + 1
	FAIL
	DISQUAL
)

func (s Status) String() string {
	switch s {
	case SUCCESS:
		return "SUCCESS"
	case FAIL:
		return "FAIL"
	case DISQUAL:
		return "DISQUAL"
	default:
		return "UNKNOWN"
	}
}
