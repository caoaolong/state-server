package orm

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// SMFlow 状态机主表：名称、描述，流程数据在 SMNode + SMEdge
type SMFlow struct {
	ID          int64          `gorm:"primaryKey"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"default:''"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime:nano"`
}

// SMNode 流程节点：NodeID 为前端 id（如 scene-start-1），Data 为 JSON（position、data 等）
type SMNode struct {
	ID        int64          `gorm:"primaryKey"`
	SMID      int64          `gorm:"not null;index"`
	NodeID    string         `gorm:"not null;size:128"` // 前端节点 id
	Type      string         `gorm:"not null;default:default"`
	Label     string         `gorm:"default:''"`
	Data      string         `gorm:"type:text;default:''"` // JSON: position, nodeCategory, nodeKind, description 等
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:nano"`
}

// SMEdge 流程边：EdgeID 为前端边 id，FromNodeID/ToNodeID 为前端节点 id
type SMEdge struct {
	ID         int64          `gorm:"primaryKey"`
	SMID       int64          `gorm:"not null;index"`
	EdgeID     string         `gorm:"not null;size:256"` // 前端边 id
	FromNodeID string         `gorm:"not null;size:128"`
	ToNodeID   string         `gorm:"not null;size:128"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	CreatedAt  time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime:nano"`
}

type SessionInfo struct {
	ID        int64          `gorm:"primaryKey"`
	SMID      int64          `gorm:"not null"`
	State     string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:nano"`
}

type SessioDetail struct {
	ID        int64          `gorm:"primaryKey"`
	SessionID int64          `gorm:"not null"`
	SMID      int64          `gorm:"not null"`
	NodeID    int64          `gorm:"not null"`
	Input     string         `gorm:"not null"`
	Output    string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime:nano"`
}

func Migrate() {
	tables := []any{
		&SMFlow{},
		&SMNode{},
		&SMEdge{},
		&SessionInfo{},
		&SessioDetail{},
	}
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log.Fatal(err)
		}
	}
}
