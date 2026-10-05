package entity

import (
	"time"
)

// BotScheduleConfig 脚本调度配置表
type BotScheduleConfig struct {
	ID                    uint      `gorm:"primaryKey;column:id"`
	Module                string    `gorm:"column:module;size:50;not null"`
	TaskName              string    `gorm:"column:task_name;size:100;not null"`
	IntervalSeconds       float64   `gorm:"column:interval_seconds;not null;default:60"`
	ExeSort               *int      `gorm:"column:exe_sort"`
	Concurrency           *int      `gorm:"column:concurrency"`
	Description           string    `gorm:"column:description;size:500"`
	IsEnabled             int       `gorm:"column:is_enabled;not null;default:1"`
	IsStrategyEnabled     *int      `gorm:"column:is_strategy_enabled;default:1"`
	LastLiveTime          string    `gorm:"column:last_live_time;size:255"`
	LastLiveTimeByMachine string    `gorm:"column:last_live_time_by_machine;type:json"`
	IsPrimaryMachineRun   int       `gorm:"column:is_primary_machine_run;not null;default:0"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

func (BotScheduleConfig) TableName() string { return "bot_schedule_config" }
