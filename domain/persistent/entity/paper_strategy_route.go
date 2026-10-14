package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperStrategyRoute struct {
	ID                   uint           `gorm:"primaryKey;column:id"`
	BehaviorKey          string         `gorm:"column:behavior_key;size:128;not null"`
	Module               string         `gorm:"column:module;size:64;not null;default:''"`
	Action               string         `gorm:"column:action;size:64;not null;default:''"`
	StrategyPackageID    uint64         `gorm:"column:strategy_package_id;not null"`
	CapabilityID         string         `gorm:"column:capability_id;size:128;not null;default:''"`
	DefaultModelConfigID uint64         `gorm:"column:default_model_config_id;not null;default:0"`
	FallbackRouteID      *uint64        `gorm:"column:fallback_route_id"`
	ConfigJSON           datatypes.JSON `gorm:"column:config_json;type:json"`
	Priority             int            `gorm:"column:priority;not null;default:0"`
	Status               string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt            time.Time      `gorm:"column:created_at"`
	UpdatedAt            time.Time      `gorm:"column:updated_at"`

	Package *PaperStrategyPackage `gorm:"foreignKey:StrategyPackageID;references:ID"`
}

func (PaperStrategyRoute) TableName() string { return "paper_strategy_route" }
