package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"testing"
	"time"
)

func TestParseIncomingEvent(t *testing.T) {
	tests := []struct {
		name      string
		line      []string
		wantEvent model.IncomingEvent
		wantErr   bool
	}{
		{
			name:      "normal event without extra param",
			line:      []string{"[14:00:00]", "1", "1"},
			wantEvent: model.IncomingEvent{PlayerID: 1, EventID: 1},
			wantErr:   false,
		},
		{
			name:      "event with extra param",
			line:      []string{"[14:44:00]", "1", "11", "50"},
			wantEvent: model.IncomingEvent{PlayerID: 1, EventID: 11, ExtraParam: "50"},
			wantErr:   false,
		},
		{name: "too short", line: []string{"[14:00:00]", "1"}, wantErr: true},
		{name: "invalid time", line: []string{"[14:00]", "1", "1"}, wantErr: true},
		{name: "invalid player id", line: []string{"[14:00:00]", "abc", "1"}, wantErr: true},
		{name: "invalid event id", line: []string{"[14:00:00]", "1", "abc"}, wantErr: true},
		{name: "empty line", line: []string{}, wantErr: true},
		{name: "invalid time 25:00:00", line: []string{"[25:00:00]", "1", "1"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIncomingEvent(tt.line)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr {
				if got.PlayerID != tt.wantEvent.PlayerID {
					t.Errorf("PlayerID: got %d, want %d", got.PlayerID, tt.wantEvent.PlayerID)
				}
				if got.EventID != tt.wantEvent.EventID {
					t.Errorf("EventID: got %d, want %d", got.EventID, tt.wantEvent.EventID)
				}
			}
		})
	}
}

func TestParseIncomingEventTimeDetails(t *testing.T) {
	line := []string{"[14:30:45]", "5", "3"}
	event, err := parseIncomingEvent(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.EventTime.Hour() != 14 {
		t.Errorf("Hour: got %d, want 14", event.EventTime.Hour())
	}
	if event.EventTime.Minute() != 30 {
		t.Errorf("Minute: got %d, want 30", event.EventTime.Minute())
	}
	if event.EventTime.Second() != 45 {
		t.Errorf("Second: got %d, want 45", event.EventTime.Second())
	}
	if event.ExtraParam != "" {
		t.Errorf("ExtraParam: got %s, want empty", event.ExtraParam)
	}
}

func TestIsDungeonOpen(t *testing.T) {
	cfg := &config.Config{OpenAt: "14:05:00", Duration: 2}
	tests := []struct {
		timeStr string
		want    bool
	}{
		{"14:04:59", false}, {"14:05:00", true}, {"15:00:00", true},
		{"16:04:59", true}, {"16:05:00", false}, {"16:05:01", false},
	}
	for _, tt := range tests {
		t.Run(tt.timeStr, func(t *testing.T) {
			eventTime, _ := time.Parse("15:04:05", tt.timeStr)
			if got := isDungeonOpen(eventTime, cfg); got != tt.want {
				t.Errorf("isDungeonOpen(%s) = %v, want %v", tt.timeStr, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d time.Duration
		s string
	}{
		{0, "00:00:00"}, {5 * time.Minute, "00:05:00"},
		{11 * time.Minute, "00:11:00"}, {24 * time.Minute, "00:24:00"},
		{1*time.Hour + 30*time.Minute, "01:30:00"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.s {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.d, got, tt.s)
		}
	}
}

func TestCalculateAverageFloorTime(t *testing.T) {
	tests := []struct {
		times []time.Duration
		want  time.Duration
	}{
		{[]time.Duration{}, 0},
		{[]time.Duration{5 * time.Minute}, 5 * time.Minute},
		{[]time.Duration{5 * time.Minute, 7 * time.Minute}, 6 * time.Minute},
	}
	for _, tt := range tests {
		player := &model.Player{FloorClearTimes: tt.times}
		if got := calculateAverageFloorTime(player); got != tt.want {
			t.Errorf("calculateAverageFloorTime(%v) = %v, want %v", tt.times, got, tt.want)
		}
	}
}

func TestCalculateBossKillTime(t *testing.T) {
	enter, _ := time.Parse("15:04:05", "14:48:00")
	kill, _ := time.Parse("15:04:05", "14:59:00")
	tests := []struct {
		player *model.Player
		want   time.Duration
	}{
		{&model.Player{BossFloorEnterTime: &enter, BossKillTime: &kill}, 11 * time.Minute},
		{&model.Player{BossFloorEnterTime: nil, BossKillTime: &kill}, 0},
		{&model.Player{BossFloorEnterTime: &enter, BossKillTime: nil}, 0},
	}
	for _, tt := range tests {
		if got := calculateBossKillTime(tt.player); got != tt.want {
			t.Errorf("calculateBossKillTime() = %v, want %v", got, tt.want)
		}
	}
}

func TestValidateAndGetPlayer(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1}

	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for non-existent player")
	}

	playInfo.Players[1] = &model.Player{ID: 1, Registered: true}
	if got := validateAndGetPlayer(event, playInfo); got == nil {
		t.Error("expected player, got nil")
	}

	playInfo.Players[1].Status = model.DISQUAL
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for disqualified player")
	}

	playInfo.Players[1].Status = model.FAIL
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for dead player")
	}
}

func TestPlayerRegistrationEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 1}

	err := playerRegistrationEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	player := playInfo.Players[1]
	if player == nil {
		t.Fatal("player not created")
	}
	if player.Health != 100 {
		t.Errorf("Health: got %d, want 100", player.Health)
	}
}

func TestPlayerGetDamageEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 100, InDungeon: true}
	eventTime, _ := time.Parse("15:04:05", "14:27:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 11, ExtraParam: "60"}

	playerGetDamageEvent(event, playInfo)
	if playInfo.Players[1].Health != 40 {
		t.Errorf("Health: got %d, want 40", playInfo.Players[1].Health)
	}

	event.ExtraParam = "50"
	playerGetDamageEvent(event, playInfo)
	if playInfo.Players[1].Health != 0 {
		t.Errorf("Health: got %d, want 0", playInfo.Players[1].Health)
	}
	if playInfo.Players[1].Status != model.FAIL {
		t.Errorf("Status: got %v, want FAIL", playInfo.Players[1].Status)
	}
}

func TestPlayerRecoveringEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 30, InDungeon: true}
	eventTime, _ := time.Parse("15:04:05", "14:49:02")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 10, ExtraParam: "80"}

	playerRecoveringEvent(event, playInfo)
	if playInfo.Players[1].Health != 100 {
		t.Errorf("Health: got %d, want 100", playInfo.Players[1].Health)
	}
}

func TestEntranceToDungeonEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 100}
	eventTime, _ := time.Parse("15:04:05", "14:10:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 2}

	entranceToTheDungeonEvent(event, playInfo)
	player := playInfo.Players[1]
	if !player.InDungeon {
		t.Error("player should be in dungeon")
	}
	if player.CurrentFloor != 1 {
		t.Errorf("CurrentFloor: got %d, want 1", player.CurrentFloor)
	}
}

func TestCalculateTimeSpent(t *testing.T) {
	enter, _ := time.Parse("15:04:05", "14:40:00")
	leave, _ := time.Parse("15:04:05", "15:04:00")
	cfg := &config.Config{OpenAt: "14:05:00", Duration: 2}

	tests := []struct {
		name   string
		player *model.Player
		want   time.Duration
	}{
		{"never entered", &model.Player{EnterTime: nil}, 0},
		{"left dungeon", &model.Player{EnterTime: &enter, LeaveTime: &leave}, 24 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTimeSpent(tt.player, cfg); got != tt.want {
				t.Errorf("calculateTimeSpent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayerWentNextFloorEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 4}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true,
		CurrentFloor: 1, FloorCleared: true,
	}
	playerWentNextFloorEvent(event, playInfo)
	if playInfo.Players[1].CurrentFloor != 2 {
		t.Errorf("CurrentFloor: got %d, want 2", playInfo.Players[1].CurrentFloor)
	}
}

func TestPlayerWentPreviousFloorEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 5}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 2,
	}
	playerWentPreviousFloorEvent(event, playInfo)
	if playInfo.Players[1].CurrentFloor != 1 {
		t.Errorf("CurrentFloor: got %d, want 1", playInfo.Players[1].CurrentFloor)
	}
}

func TestPlayerEnteredBossFloorEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 6}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 2,
	}
	playerEnteredBossFloorEvent(event, playInfo)
	if playInfo.Players[1].BossFloorEnterTime == nil {
		t.Error("BossFloorEnterTime should be set")
	}
}

func TestPlayerKilledBossEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:59:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 7}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 2,
	}
	playerKilledBossEvent(event, playInfo)
	if !playInfo.Players[1].BossDefeated {
		t.Error("BossDefeated should be true")
	}
	if playInfo.Players[1].Status != model.SUCCESS {
		t.Errorf("Status: got %v, want SUCCESS", playInfo.Players[1].Status)
	}
}

func TestPlayerLeftDungeonEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "15:04:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 8}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true,
	}
	playerLeftDungeonEvent(event, playInfo)
	if playInfo.Players[1].InDungeon {
		t.Error("InDungeon should be false")
	}
	if playInfo.Players[1].LeaveTime == nil {
		t.Error("LeaveTime should be set")
	}
}

func TestPlayerCannotContinue(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:10:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 9}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true,
	}
	playerCannotContinue(event, playInfo)
	if playInfo.Players[1].Status != model.DISQUAL {
		t.Errorf("Status: got %v, want DISQUAL", playInfo.Players[1].Status)
	}
}

func TestFinalReport(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	enter, _ := time.Parse("15:04:05", "14:40:00")
	leave, _ := time.Parse("15:04:05", "15:04:00")

	playInfo.Players[1] = &model.Player{
		ID: 1, Health: 35, Status: model.SUCCESS,
		EnterTime: &enter, LeaveTime: &leave,
		FloorClearTimes: []time.Duration{5 * time.Minute},
	}
	err := FinalReport(playInfo)
	if err != nil {
		t.Errorf("FinalReport error: %v", err)
	}
}

func TestPlayerKilledMonsterEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:41:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 3}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true,
		CurrentFloor: 1, FloorCleared: false, KilledOnCurrent: 1,
		CurrentFloorEnterTime: &eventTime,
	}
	playerKilledMonsterEvent(event, playInfo)
	if playInfo.Players[1].KilledOnCurrent != 2 {
		t.Errorf("KilledOnCurrent: got %d, want 2", playInfo.Players[1].KilledOnCurrent)
	}
	if !playInfo.Players[1].FloorCleared {
		t.Error("FloorCleared should be true after killing last monster")
	}
	if len(playInfo.Players[1].FloorClearTimes) != 1 {
		t.Errorf("FloorClearTimes len: got %d, want 1", len(playInfo.Players[1].FloorClearTimes))
	}
}

func TestPlayerKilledMonsterEventInvalid(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:41:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 3}

	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, InDungeon: false}
	playerKilledMonsterEvent(event, playInfo)

	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, InDungeon: true, CurrentFloor: 2}
	playerKilledMonsterEvent(event, playInfo)

	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, InDungeon: true, CurrentFloor: 1, FloorCleared: true}
	playerKilledMonsterEvent(event, playInfo)
}

func TestEventHandlerAllCases(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")

	testCases := []struct {
		name    string
		eventID int
		wantErr bool
	}{
		{"event 1", 1, false},
		{"event 2", 2, false},
		{"event 3", 3, false},
		{"event 4", 4, false},
		{"event 5", 5, false},
		{"event 6", 6, false},
		{"event 7", 7, false},
		{"event 8", 8, false},
		{"event 9", 9, false},
		{"event 10", 10, false},
		{"event 11", 11, false},
		{"unknown event", 99, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: tc.eventID}
			if tc.eventID == 10 || tc.eventID == 11 {
				event.ExtraParam = "10"
			}
			err := eventHandler(event, playInfo)
			if tc.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPlayerRegistrationEventInvalidID(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 0, EventID: 1}

	err := playerRegistrationEvent(event, playInfo)
	if err == nil {
		t.Error("expected error for invalid player id")
	}
}

func TestPlayerRecoveringEventOverheal(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 90, InDungeon: true}
	eventTime, _ := time.Parse("15:04:05", "14:49:02")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 10, ExtraParam: "20"}

	playerRecoveringEvent(event, playInfo)
	if playInfo.Players[1].Health != 100 {
		t.Errorf("Health: got %d, want 100 (capped)", playInfo.Players[1].Health)
	}
}

func TestPlayerGetDamageEventDeath(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 30, InDungeon: true}
	eventTime, _ := time.Parse("15:04:05", "14:27:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 11, ExtraParam: "50"}

	playerGetDamageEvent(event, playInfo)
	if playInfo.Players[1].Health != 0 {
		t.Errorf("Health: got %d, want 0", playInfo.Players[1].Health)
	}
	if playInfo.Players[1].Status != model.FAIL {
		t.Errorf("Status: got %v, want FAIL", playInfo.Players[1].Status)
	}
}

func TestEntranceToDungeonEventAlreadyInside(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:10:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 2}

	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, InDungeon: true}
	entranceToTheDungeonEvent(event, playInfo)
}

func TestPlayerWentNextFloorEventNotCleared(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 4}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true,
		CurrentFloor: 1, FloorCleared: false,
	}
	playerWentNextFloorEvent(event, playInfo)
}

func TestPlayerWentPreviousFloorEventFirstFloor(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 5}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 1,
	}
	playerWentPreviousFloorEvent(event, playInfo)
}

func TestPlayerEnteredBossFloorEventNotInDungeon(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 6}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: false,
	}
	playerEnteredBossFloorEvent(event, playInfo)
}

func TestPlayerKilledBossEventWrongFloor(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:59:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 7}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 1,
	}
	playerKilledBossEvent(event, playInfo)
}

func TestPlayerLeftDungeonEventNotInDungeon(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "15:04:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 8}

	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: false,
	}
	playerLeftDungeonEvent(event, playInfo)
}
