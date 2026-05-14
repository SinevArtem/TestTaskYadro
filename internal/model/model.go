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
	ID              int
	Registered      bool
	InDungeon       bool            // в подземелье
	CurrentFloor    int             // 1-based, этаж босса = Floors+1
	KilledOnCurrent int             // сколько монстров убито на текущем этаже
	Health          int             // 0-100
	EnterTime       *time.Time      // когда вошёл в подземелье
	FloorClearTimes []time.Duration // время зачистки каждого обычного этажа
	BossKillTime    *time.Time      // когда убил босса
	LeaveTime       *time.Time      // когда покинул
	Disqualified    bool
	Dead            bool
	CannotContinue  bool
	ImpossibleMove  bool //Невозможный ход
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
