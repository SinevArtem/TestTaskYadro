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
			wantEvent: model.IncomingEvent{PlayerID: 1, EventID: 1, ExtraParam: ""},
			wantErr:   false,
		},
		{
			name:      "event with extra param",
			line:      []string{"[14:44:00]", "1", "11", "50"},
			wantEvent: model.IncomingEvent{PlayerID: 1, EventID: 11, ExtraParam: "50"},
			wantErr:   false,
		},
		{
			name:      "event with multi-word extra param",
			line:      []string{"[14:44:00]", "1", "9", "connection", "lost"},
			wantEvent: model.IncomingEvent{PlayerID: 1, EventID: 9, ExtraParam: "connection"},
			wantErr:   false,
		},
		{name: "too short", line: []string{"[14:00:00]", "1"}, wantErr: true},
		{name: "invalid time format", line: []string{"[14:00]", "1", "1"}, wantErr: true},
		{name: "invalid player id", line: []string{"[14:00:00]", "abc", "1"}, wantErr: true},
		{name: "invalid event id", line: []string{"[14:00:00]", "1", "abc"}, wantErr: true},
		{name: "empty line", line: []string{}, wantErr: true},
		{name: "invalid time 25:00:00", line: []string{"[25:00:00]", "1", "1"}, wantErr: true},
		{name: "negative player id", line: []string{"[14:00:00]", "-1", "1"}, wantErr: true},
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
				if tt.wantEvent.ExtraParam != "" && got.ExtraParam != tt.wantEvent.ExtraParam {
					t.Errorf("ExtraParam: got %s, want %s", got.ExtraParam, tt.wantEvent.ExtraParam)
				}
			}
		})
	}
}

func TestIsDungeonOpen(t *testing.T) {
	cfg := &config.Config{OpenAt: "14:05:00", Duration: 2}
	tests := []struct {
		timeStr string
		want    bool
	}{
		{"14:04:59", false},
		{"14:05:00", true},
		{"15:00:00", true},
		{"16:04:59", true},
		{"16:05:00", false},
		{"16:05:01", false},
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

func TestValidateAndGetPlayer(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1}

	// Test non-existent player
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for non-existent player")
	}

	// Test registered player
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 100}
	if got := validateAndGetPlayer(event, playInfo); got == nil {
		t.Error("expected player, got nil")
	}

	// Test disqualified player
	playInfo.Players[1].Status = model.DISQUAL
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for disqualified player")
	}

	// Test dead player
	playInfo.Players[1].Status = model.FAIL
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for dead player")
	}

	// Test unregistered player
	playInfo.Players[2] = &model.Player{ID: 2, Registered: false}
	event.PlayerID = 2
	if got := validateAndGetPlayer(event, playInfo); got != nil {
		t.Error("expected nil for unregistered player")
	}
}

func TestPlayerRegistrationEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")

	// Test valid registration
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
	if !player.Registered {
		t.Error("Player should be registered")
	}
	if player.Status != 0 {
		t.Errorf("Status: got %v, want 0", player.Status)
	}

	// Test invalid player ID (0 or negative)
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 0, EventID: 1}
	err = playerRegistrationEvent(event, playInfo)
	if err == nil {
		t.Error("expected error for invalid player id (0)")
	}
}

func TestPlayerGetDamageEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:27:00")

	// Create valid player
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 100, InDungeon: true}

	// Test normal damage
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 11, ExtraParam: "60"}
	err := playerGetDamageEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if playInfo.Players[1].Health != 40 {
		t.Errorf("Health: got %d, want 40", playInfo.Players[1].Health)
	}

	// Test fatal damage
	event.ExtraParam = "50"
	err = playerGetDamageEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if playInfo.Players[1].Health != 0 {
		t.Errorf("Health: got %d, want 0", playInfo.Players[1].Health)
	}
	if playInfo.Players[1].Status != model.FAIL {
		t.Errorf("Status: got %v, want FAIL", playInfo.Players[1].Status)
	}
	if playInfo.Players[1].LeaveTime == nil {
		t.Error("LeaveTime should be set when player dies")
	}

	playInfo.Players[2] = &model.Player{ID: 2, Registered: true, Health: 100, InDungeon: true}
	event.PlayerID = 2
	event.ExtraParam = "invalid"
	err = playerGetDamageEvent(event, playInfo)
	if err == nil {
		t.Error("expected error for invalid damage value")
	}
}

func TestPlayerRecoveringEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:49:02")

	// Create valid player
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 30, InDungeon: true}

	// Test normal heal
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 10, ExtraParam: "50"}
	err := playerRecoveringEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if playInfo.Players[1].Health != 80 {
		t.Errorf("Health: got %d, want 80", playInfo.Players[1].Health)
	}

	// Test overheal (capped at 100)
	event.ExtraParam = "80"
	err = playerRecoveringEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if playInfo.Players[1].Health != 100 {
		t.Errorf("Health: got %d, want 100 (capped)", playInfo.Players[1].Health)
	}

	// Test invalid heal param - should return error
	event.ExtraParam = "invalid"
	err = playerRecoveringEvent(event, playInfo)
	if err == nil {
		t.Error("expected error for invalid heal value")
	}
}

func TestEntranceToTheDungeonEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:10:00")

	// Test successful entrance
	playInfo.Players[1] = &model.Player{ID: 1, Registered: true, Health: 100}
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 2}
	entranceToTheDungeonEvent(event, playInfo)
	player := playInfo.Players[1]
	if !player.InDungeon {
		t.Error("player should be in dungeon")
	}
	if player.CurrentFloor != 1 {
		t.Errorf("CurrentFloor: got %d, want 1", player.CurrentFloor)
	}
	if player.EnterTime == nil {
		t.Error("EnterTime should be set")
	}
	if player.CurrentFloorEnterTime == nil {
		t.Error("CurrentFloorEnterTime should be set")
	}

	// Test entrance when already inside - state should not change
	originalFloor := player.CurrentFloor
	entranceToTheDungeonEvent(event, playInfo)
	if player.CurrentFloor != originalFloor {
		t.Errorf("CurrentFloor changed from %d to %d", originalFloor, player.CurrentFloor)
	}
}

func TestPlayerKilledMonsterEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:41:00")

	playInfo.Players[1] = &model.Player{
		ID:                    1,
		Registered:            true,
		InDungeon:             true,
		CurrentFloor:          1,
		FloorCleared:          false,
		KilledOnCurrent:       1,
		CurrentFloorEnterTime: &eventTime,
	}

	// Kill second monster (should complete floor)
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 3}
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
	if playInfo.Players[1].KilledOnCurrent != 0 {
		t.Error("KilledOnCurrent should be reset")
	}
	if playInfo.Players[1].FloorCleared {
		t.Error("FloorCleared should be false on new floor")
	}

	playInfo.Players[2] = &model.Player{
		ID: 2, Registered: true, InDungeon: true,
		CurrentFloor: 1, FloorCleared: false,
	}
	event.PlayerID = 2
	originalFloor := playInfo.Players[2].CurrentFloor
	playerWentNextFloorEvent(event, playInfo)
	if playInfo.Players[2].CurrentFloor != originalFloor {
		t.Errorf("CurrentFloor changed from %d to %d", originalFloor, playInfo.Players[2].CurrentFloor)
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
	if !playInfo.Players[1].FloorCleared {
		t.Error("FloorCleared should be true when going back")
	}

	// Test previous floor on first floor - should not change
	playInfo.Players[2] = &model.Player{
		ID: 2, Registered: true, InDungeon: true, CurrentFloor: 1,
	}
	event.PlayerID = 2
	originalFloor := playInfo.Players[2].CurrentFloor
	playerWentPreviousFloorEvent(event, playInfo)
	if playInfo.Players[2].CurrentFloor != originalFloor {
		t.Errorf("CurrentFloor changed from %d to %d", originalFloor, playInfo.Players[2].CurrentFloor)
	}
}

func TestPlayerEnteredBossFloorEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:48:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 6}

	// Test successful boss floor entry
	playInfo.Players[1] = &model.Player{
		ID: 1, Registered: true, InDungeon: true, CurrentFloor: 2,
	}
	playerEnteredBossFloorEvent(event, playInfo)
	if playInfo.Players[1].BossFloorEnterTime == nil {
		t.Error("BossFloorEnterTime should be set")
	}

	// Test boss floor entry on wrong floor - should not set
	playInfo.Players[2] = &model.Player{
		ID: 2, Registered: true, InDungeon: true, CurrentFloor: 1,
	}
	event.PlayerID = 2
	playerEnteredBossFloorEvent(event, playInfo)
	if playInfo.Players[2].BossFloorEnterTime != nil {
		t.Error("BossFloorEnterTime should not be set on wrong floor")
	}
}

func TestPlayerKilledBossEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:59:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 7}

	// Test successful boss kill
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
	if playInfo.Players[1].BossKillTime == nil {
		t.Error("BossKillTime should be set")
	}

	// Test boss kill on wrong floor - should not succeed
	playInfo.Players[2] = &model.Player{
		ID: 2, Registered: true, InDungeon: true, CurrentFloor: 1,
	}
	event.PlayerID = 2
	playerKilledBossEvent(event, playInfo)
	if playInfo.Players[2].BossDefeated {
		t.Error("BossDefeated should not be true on wrong floor")
	}
}

func TestPlayerLeftDungeonEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "15:04:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 8}

	// Test successful leave
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

	playInfo.Players[2] = &model.Player{
		ID: 2, Registered: true, InDungeon: false, LeaveTime: nil,
	}
	event.PlayerID = 2
	playerLeftDungeonEvent(event, playInfo)
	if playInfo.Players[2].LeaveTime != nil {
		t.Error("LeaveTime should not be set when not in dungeon")
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
	if playInfo.Players[1].LeaveTime == nil {
		t.Error("LeaveTime should be set")
	}

	event.PlayerID = 2
	playerCannotContinue(event, playInfo)
	if playInfo.Players[2] == nil {
		t.Error("Player should be created for unregistered")
	}
	if playInfo.Players[2].Status != model.DISQUAL {
		t.Errorf("Status: got %v, want DISQUAL", playInfo.Players[2].Status)
	}
}

func TestFinalReport(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	enter, _ := time.Parse("15:04:05", "14:40:00")
	leave, _ := time.Parse("15:04:05", "15:04:00")
	bossEnter, _ := time.Parse("15:04:05", "14:48:00")
	bossKill, _ := time.Parse("15:04:05", "14:59:00")

	// Test SUCCESS player
	playInfo.Players[1] = &model.Player{
		ID: 1, Health: 35, Status: model.SUCCESS,
		EnterTime: &enter, LeaveTime: &leave,
		FloorClearTimes:    []time.Duration{5 * time.Minute},
		BossFloorEnterTime: &bossEnter,
		BossKillTime:       &bossKill,
		BossDefeated:       true,
	}

	// Test FAIL player
	playInfo.Players[2] = &model.Player{
		ID: 2, Health: 0, Status: model.FAIL,
		EnterTime: &enter, LeaveTime: &leave,
	}

	// Test DISQUAL player (never entered)
	playInfo.Players[3] = &model.Player{
		ID: 3, Health: 100, Status: model.DISQUAL, Registered: true,
	}

	err := FinalReport(playInfo)
	if err != nil {
		t.Errorf("FinalReport error: %v", err)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d time.Duration
		s string
	}{
		{0, "00:00:00"},
		{5 * time.Minute, "00:05:00"},
		{11 * time.Minute, "00:11:00"},
		{24 * time.Minute, "00:24:00"},
		{1*time.Hour + 30*time.Minute, "01:30:00"},
		{1*time.Hour + 30*time.Minute + 45*time.Second, "01:30:45"},
		{25*time.Hour + 5*time.Minute, "25:05:00"},
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
		{[]time.Duration{5 * time.Minute, 7 * time.Minute, 9 * time.Minute}, 7 * time.Minute},
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
		name   string
		player *model.Player
		want   time.Duration
	}{
		{"normal boss kill", &model.Player{BossFloorEnterTime: &enter, BossKillTime: &kill}, 11 * time.Minute},
		{"no enter time", &model.Player{BossFloorEnterTime: nil, BossKillTime: &kill}, 0},
		{"no kill time", &model.Player{BossFloorEnterTime: &enter, BossKillTime: nil}, 0},
		{"both nil", &model.Player{BossFloorEnterTime: nil, BossKillTime: nil}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateBossKillTime(tt.player); got != tt.want {
				t.Errorf("calculateBossKillTime() = %v, want %v", got, tt.want)
			}
		})
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

func TestEventHandlerAllCases(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}
	eventTime, _ := time.Parse("15:04:05", "14:00:00")

	// Register player first
	regEvent := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 1}
	playerRegistrationEvent(regEvent, playInfo)

	// Also add player to dungeon for movement events
	playInfo.Players[1].InDungeon = true
	playInfo.Players[1].CurrentFloor = 1
	playInfo.Players[1].FloorCleared = true

	testCases := []struct {
		name    string
		eventID int
		wantErr bool
	}{
		{"event 1 - registration", 1, false},
		{"event 2 - enter dungeon", 2, false},
		{"event 3 - kill monster", 3, false},
		{"event 4 - next floor", 4, false},
		{"event 5 - previous floor", 5, false},
		{"event 6 - enter boss floor", 6, false},
		{"event 7 - kill boss", 7, false},
		{"event 8 - leave dungeon", 8, false},
		{"event 9 - cannot continue", 9, false},
		{"event 10 - restore health", 10, false},
		{"event 11 - receive damage", 11, false},
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

// Test complex scenario from the example
func TestComplexScenario(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config:  &config.Config{Floors: 2, Monsters: 2, OpenAt: "14:05:00", Duration: 2},
	}

	// Player 1 registration
	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 1}
	playerRegistrationEvent(event, playInfo)

	// Player 1 enters dungeon (after open time)
	eventTime, _ = time.Parse("15:04:05", "14:40:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 2}
	entranceToTheDungeonEvent(event, playInfo)

	// Kill monsters on floor 1
	eventTime, _ = time.Parse("15:04:05", "14:41:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 3}
	playerKilledMonsterEvent(event, playInfo)

	eventTime, _ = time.Parse("15:04:05", "14:45:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 3}
	playerKilledMonsterEvent(event, playInfo)

	// Go to next floor
	eventTime, _ = time.Parse("15:04:05", "14:48:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 4}
	playerWentNextFloorEvent(event, playInfo)

	// Enter boss floor
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 6}
	playerEnteredBossFloorEvent(event, playInfo)

	// Take damage and heal
	eventTime, _ = time.Parse("15:04:05", "14:49:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 11, ExtraParam: "25"}
	playerGetDamageEvent(event, playInfo)

	eventTime, _ = time.Parse("15:04:05", "14:49:02")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 10, ExtraParam: "80"}
	playerRecoveringEvent(event, playInfo)

	eventTime, _ = time.Parse("15:04:05", "14:50:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 11, ExtraParam: "65"}
	playerGetDamageEvent(event, playInfo)

	// Kill boss
	eventTime, _ = time.Parse("15:04:05", "14:59:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 7}
	playerKilledBossEvent(event, playInfo)

	// Leave dungeon
	eventTime, _ = time.Parse("15:04:05", "15:04:00")
	event = model.IncomingEvent{EventTime: eventTime, PlayerID: 1, EventID: 8}
	playerLeftDungeonEvent(event, playInfo)

	// Verify final state
	player := playInfo.Players[1]
	if player.Status != model.SUCCESS {
		t.Errorf("Expected SUCCESS, got %v", player.Status)
	}
	if player.Health != 35 {
		t.Errorf("Expected health 35, got %d", player.Health)
	}
	if len(player.FloorClearTimes) != 1 {
		t.Errorf("Expected 1 floor clear time, got %d", len(player.FloorClearTimes))
	}
	if !player.BossDefeated {
		t.Error("Boss should be defeated")
	}
}
