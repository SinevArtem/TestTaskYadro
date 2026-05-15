package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0 * time.Second, "00:00:00"},
		{"5 minutes", 5 * time.Minute, "00:05:00"},
		{"11 minutes", 11 * time.Minute, "00:11:00"},
		{"19 minutes", 19 * time.Minute, "00:19:00"},
		{"24 minutes", 24 * time.Minute, "00:24:00"},
		{"1 hour 30 minutes", 1*time.Hour + 30*time.Minute, "01:30:00"},
		{"2 hours 5 minutes 10 seconds", 2*time.Hour + 5*time.Minute + 10*time.Second, "02:05:10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestCalculateTimeSpent(t *testing.T) {
	enterTime, _ := time.Parse("15:04:05", "14:40:00")

	cfg := &config.Config{
		OpenAt:   "14:05:00",
		Duration: 2,
	}

	tests := []struct {
		name     string
		player   *model.Player
		expected time.Duration
	}{
		{
			name: "player left the dungeon",
			player: &model.Player{
				EnterTime: &enterTime,
				LeaveTime: func() *time.Time {
					t, _ := time.Parse("15:04:05", "15:04:00")
					return &t
				}(),
			},
			expected: 24 * time.Minute,
		},
		{
			name: "player still in dungeon (use close time)",
			player: &model.Player{
				EnterTime: &enterTime,
				LeaveTime: nil,
			},
			expected: 0,
		},
		{
			name: "player never entered",
			player: &model.Player{
				EnterTime: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTimeSpent(tt.player, cfg)
			if tt.name == "player still in dungeon (use close time)" {
				if result == 0 {
					t.Error("expected non-zero duration for player in dungeon")
				}
			} else if result != tt.expected {
				t.Errorf("calculateTimeSpent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateAverageFloorTime(t *testing.T) {
	tests := []struct {
		name     string
		player   *model.Player
		expected time.Duration
	}{
		{
			name: "no floors cleared",
			player: &model.Player{
				FloorClearTimes: []time.Duration{},
			},
			expected: 0,
		},
		{
			name: "one floor cleared in 5 minutes",
			player: &model.Player{
				FloorClearTimes: []time.Duration{5 * time.Minute},
			},
			expected: 5 * time.Minute,
		},
		{
			name: "two floors cleared",
			player: &model.Player{
				FloorClearTimes: []time.Duration{5 * time.Minute, 7 * time.Minute},
			},
			expected: 6 * time.Minute,
		},
		{
			name: "three floors cleared",
			player: &model.Player{
				FloorClearTimes: []time.Duration{5 * time.Minute, 7 * time.Minute, 4 * time.Minute},
			},
			expected: (5 + 7 + 4) * time.Minute / 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAverageFloorTime(tt.player)
			if result != tt.expected {
				t.Errorf("calculateAverageFloorTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateBossKillTime(t *testing.T) {
	enterTime, _ := time.Parse("15:04:05", "14:48:00")
	killTime, _ := time.Parse("15:04:05", "14:59:00")

	tests := []struct {
		name     string
		player   *model.Player
		expected time.Duration
	}{
		{
			name: "boss killed",
			player: &model.Player{
				BossFloorEnterTime: &enterTime,
				BossKillTime:       &killTime,
			},
			expected: 11 * time.Minute,
		},
		{
			name: "boss not killed - no kill time",
			player: &model.Player{
				BossFloorEnterTime: &enterTime,
				BossKillTime:       nil,
			},
			expected: 0,
		},
		{
			name: "boss not killed - no enter time",
			player: &model.Player{
				BossFloorEnterTime: nil,
				BossKillTime:       &killTime,
			},
			expected: 0,
		},
		{
			name: "boss not killed - both nil",
			player: &model.Player{
				BossFloorEnterTime: nil,
				BossKillTime:       nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBossKillTime(tt.player)
			if result != tt.expected {
				t.Errorf("calculateBossKillTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateTimeSpentWithCloseTime(t *testing.T) {
	enterTime, _ := time.Parse("15:04:05", "14:40:00")

	cfg := &config.Config{
		OpenAt:   "14:05:00",
		Duration: 2,
	}

	player := &model.Player{
		EnterTime: &enterTime,
		LeaveTime: nil,
	}

	result := calculateTimeSpent(player, cfg)

	// openTime = 14:05, closeTime = 16:05
	// enterTime = 14:40
	// expected = 16:05 - 14:40 = 1 hour 25 minutes = 85 minutes
	expected := 85 * time.Minute

	if result != expected {
		t.Errorf("calculateTimeSpent() = %v, want %v", result, expected)
	}
}

func TestFinalReportFormat(t *testing.T) {
	playInfo := &model.PlayInfo{
		Players: make(map[int]*model.Player),
		Config: &config.Config{
			Floors:   2,
			Monsters: 2,
			OpenAt:   "14:05:00",
			Duration: 2,
		},
	}

	enterTime, _ := time.Parse("15:04:05", "14:40:00")
	leaveTime, _ := time.Parse("15:04:05", "15:04:00")

	playInfo.Players[1] = &model.Player{
		ID:                 1,
		Health:             35,
		Status:             model.SUCCESS,
		EnterTime:          &enterTime,
		LeaveTime:          &leaveTime,
		FloorClearTimes:    []time.Duration{5 * time.Minute},
		BossFloorEnterTime: &enterTime,
		BossKillTime: func() *time.Time {
			t, _ := time.Parse("15:04:05", "14:59:00")
			return &t
		}(),
	}

	err := FinalReport(playInfo)
	if err != nil {
		t.Errorf("FinalReport returned error: %v", err)
	}
}
