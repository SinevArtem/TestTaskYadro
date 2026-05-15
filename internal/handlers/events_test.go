package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"testing"
	"time"
)

func TestPlayerRegistrationEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{
		EventTime: eventTime,
		PlayerID:  1,
		EventID:   1,
	}

	err := playerRegistrationEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	player, ok := playInfo.Players[1]
	if !ok {
		t.Fatal("player not created")
	}

	if player.ID != 1 {
		t.Errorf("ID: got %d, want 1", player.ID)
	}
	if player.Health != 100 {
		t.Errorf("Health: got %d, want 100", player.Health)
	}
	if !player.Registered {
		t.Error("Registered should be true")
	}
}

func TestPlayerGetDamageEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	playInfo.Players[1] = &model.Player{
		ID:         1,
		Registered: true,
		Health:     100,
		InDungeon:  true,
	}

	eventTime, _ := time.Parse("15:04:05", "14:27:00")
	event := model.IncomingEvent{
		EventTime:  eventTime,
		PlayerID:   1,
		EventID:    11,
		ExtraParam: "60",
	}

	err := playerGetDamageEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	player := playInfo.Players[1]
	if player.Health != 40 {
		t.Errorf("Health: got %d, want 40", player.Health)
	}

	event.ExtraParam = "50"
	err = playerGetDamageEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if player.Health != 0 {
		t.Errorf("Health: got %d, want 0", player.Health)
	}
	if player.Status != model.FAIL {
		t.Errorf("Status: got %v, want FAIL", player.Status)
	}
}

func TestPlayerRecoveringEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	playInfo.Players[1] = &model.Player{
		ID:         1,
		Registered: true,
		Health:     30,
		InDungeon:  true,
	}

	eventTime, _ := time.Parse("15:04:05", "14:49:02")
	event := model.IncomingEvent{
		EventTime:  eventTime,
		PlayerID:   1,
		EventID:    10,
		ExtraParam: "80",
	}

	err := playerRecoveringEvent(event, playInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	player := playInfo.Players[1]
	if player.Health != 100 {
		t.Errorf("Health: got %d, want 100", player.Health)
	}
}

func TestIsDungeonOpen(t *testing.T) {
	cfg := &config.Config{
		OpenAt:   "14:05:00",
		Duration: 2,
	}

	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"before open", "14:04:59", false},
		{"at open", "14:05:00", true},
		{"during open", "15:00:00", true},
		{"before close", "16:04:59", true},
		{"at close", "16:05:00", false},
		{"after close", "16:05:01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventTime, _ := time.Parse("15:04:05", tt.timeStr)
			result := isDungeonOpen(eventTime, cfg)
			if result != tt.expected {
				t.Errorf("isDungeonOpen(%s) = %v, want %v", tt.timeStr, result, tt.expected)
			}
		})
	}
}

func TestValidateAndGetPlayer(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	eventTime, _ := time.Parse("15:04:05", "14:00:00")
	event := model.IncomingEvent{
		EventTime: eventTime,
		PlayerID:  1,
		EventID:   2,
	}

	player := validateAndGetPlayer(event, playInfo)
	if player != nil {
		t.Error("expected nil for non-existent player")
	}

	playInfo.Players[1] = &model.Player{
		ID:         1,
		Registered: true,
	}

	player = validateAndGetPlayer(event, playInfo)
	if player == nil {
		t.Error("expected player, got nil")
	}

	playInfo.Players[1].Status = model.DISQUAL
	player = validateAndGetPlayer(event, playInfo)
	if player != nil {
		t.Error("expected nil for disqualified player")
	}

	playInfo.Players[1].Status = model.FAIL
	player = validateAndGetPlayer(event, playInfo)
	if player != nil {
		t.Error("expected nil for dead player")
	}
}

func TestEntranceToDungeonEvent(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	playInfo.Players[1] = &model.Player{
		ID:         1,
		Registered: true,
		Health:     100,
	}

	eventTime, _ := time.Parse("15:04:05", "14:10:00")
	event := model.IncomingEvent{
		EventTime: eventTime,
		PlayerID:  1,
		EventID:   2,
	}

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
}
