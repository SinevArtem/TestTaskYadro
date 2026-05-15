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
			name: "normal event without extra param",
			line: []string{"[14:00:00]", "1", "1"},
			wantEvent: model.IncomingEvent{
				PlayerID: 1,
				EventID:  1,
			},
			wantErr: false,
		},
		{
			name: "event with extra param",
			line: []string{"[14:44:00]", "1", "11", "50"},
			wantEvent: model.IncomingEvent{
				PlayerID:   1,
				EventID:    11,
				ExtraParam: "50",
			},
			wantErr: false,
		},
		{
			name:    "invalid line - too short",
			line:    []string{"[14:00:00]", "1"},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			line:    []string{"[14:00]", "1", "1"},
			wantErr: true,
		},
		{
			name:    "invalid player id",
			line:    []string{"[14:00:00]", "abc", "1"},
			wantErr: true,
		},
		{
			name:    "invalid event id",
			line:    []string{"[14:00:00]", "1", "abc"},
			wantErr: true,
		},
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
				if got.ExtraParam != tt.wantEvent.ExtraParam {
					t.Errorf("ExtraParam: got %s, want %s", got.ExtraParam, tt.wantEvent.ExtraParam)
				}
				if got.EventTime.Hour() != 14 {
					t.Errorf("EventTime hour: got %d, want 14", got.EventTime.Hour())
				}
			}
		})
	}
}

func TestEventHandler(t *testing.T) {
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

	tests := []struct {
		name    string
		event   model.IncomingEvent
		wantErr bool
	}{
		{
			name: "event 1 - registration",
			event: model.IncomingEvent{
				EventTime: eventTime,
				PlayerID:  1,
				EventID:   1,
			},
			wantErr: false,
		},
		{
			name: "event 2 - entrance",
			event: model.IncomingEvent{
				EventTime: eventTime,
				PlayerID:  1,
				EventID:   2,
			},
			wantErr: false,
		},
		{
			name: "unknown event",
			event: model.IncomingEvent{
				EventTime: eventTime,
				PlayerID:  1,
				EventID:   99,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eventHandler(tt.event, playInfo)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseIncomingEventTime(t *testing.T) {
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
	if event.PlayerID != 5 {
		t.Errorf("PlayerID: got %d, want 5", event.PlayerID)
	}
	if event.EventID != 3 {
		t.Errorf("EventID: got %d, want 3", event.EventID)
	}
}

func TestParseIncomingEventWithExtraParam(t *testing.T) {
	line := []string{"[14:30:45]", "5", "11", "75"}

	event, err := parseIncomingEvent(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.ExtraParam != "75" {
		t.Errorf("ExtraParam: got %s, want 75", event.ExtraParam)
	}
}

func TestParseIncomingEventEmptyLine(t *testing.T) {
	line := []string{}

	_, err := parseIncomingEvent(line)
	if err == nil {
		t.Error("expected error for empty line")
	}
}

func TestParseIncomingEventInvalidTime(t *testing.T) {
	line := []string{"[25:00:00]", "1", "1"}

	_, err := parseIncomingEvent(line)
	if err == nil {
		t.Error("expected error for invalid time")
	}
}
