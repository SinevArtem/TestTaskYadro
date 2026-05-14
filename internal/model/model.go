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
	Disqualified          bool
	Dead                  bool
}

type IncomingEvent struct {
	EventTime  time.Time
	PlayerID   int
	EventID    int
	ExtraParam string
}

type Event struct {
	ID         int
	ExtraParam *int
	Comment    string
}

type PlayerReport struct {
	State             string
	PlayerID          int
	TimeInTheDungeon  time.Time // время проведенное в подземелье
	AverageFloorTime  time.Time // среднее время этажа
	TimeToKillTheBoss time.Time // вермя на босса
	Health            int
}
