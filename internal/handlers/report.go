package handlers

import (
	"TestTaskYadro/internal/config"
	"TestTaskYadro/internal/model"
	"fmt"
	"time"
)

func FinalReport(playInfo *model.PlayInfo) error {
	for _, player := range playInfo.Players {
		if player.Status == 0 {
			if player.EnterTime == nil {
				player.Status = model.DISQUAL
			} else if !player.BossDefeated {
				player.Status = model.FAIL
			}
		}
	}

	fmt.Println("Final report:")
	for _, player := range playInfo.Players {

		timeSpent := calculateTimeSpent(player, playInfo.Config)
		avgTime := calculateAverageFloorTime(player)
		bossTime := calculateBossKillTime(player)

		fmt.Printf("[%s] %d [%s, %s, %s] HP:%d\n",
			player.Status.String(),
			player.ID,
			formatDuration(timeSpent),
			formatDuration(avgTime),
			formatDuration(bossTime),
			player.Health)
	}
	return nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func calculateTimeSpent(player *model.Player, cfg *config.Config) time.Duration {
	if player.EnterTime == nil {
		return 0
	}

	endTime := player.LeaveTime
	if endTime == nil {
		openTime, _ := time.Parse("15:04:05", cfg.OpenAt)
		closeTime := openTime.Add(time.Duration(cfg.Duration) * time.Hour)
		endTime = &closeTime
	}

	return endTime.Sub(*player.EnterTime)
}

func calculateAverageFloorTime(player *model.Player) time.Duration {
	if len(player.FloorClearTimes) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, t := range player.FloorClearTimes {
		total += t
	}
	return total / time.Duration(len(player.FloorClearTimes))
}

func calculateBossKillTime(player *model.Player) time.Duration {
	if player.BossKillTime == nil || player.BossFloorEnterTime == nil {
		return 0
	}
	return player.BossKillTime.Sub(*player.BossFloorEnterTime)
}
