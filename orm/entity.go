package orm

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// SMFlow 状态机主表：名称、描述、base_url，流程数据在 SMNode + SMEdge
type SMFlow struct {
	ID          int64          `gorm:"primaryKey"`
	Identifier  string         `gorm:"default:''"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"default:''"`
	BaseURL     string         `gorm:"default:''"` // 请求基础地址，与节点请求路径拼接
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime:nano"`
}

// SMApiKey ApiKey 管理表，表名 sm_apikey
type SMApiKey struct {
	ID        int64          `gorm:"primaryKey"`
	Name      string         `gorm:"not null;size:128"` // 名称/备注
	ApiKey    string         `gorm:"not null;size:256"` // 密钥
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:nano"`
}

func (SMApiKey) TableName() string { return "sm_apikey" }

// SMNode 流程节点：NodeID 为前端 id，Data 为 JSON；请求地址不包含 base_url
type SMNode struct {
	ID            int64          `gorm:"primaryKey"`
	SMID          int64          `gorm:"not null;index"`
	NodeID        string         `gorm:"not null;size:128"` // 前端节点 id
	Type          string         `gorm:"not null;default:default"`
	Label         string         `gorm:"default:''"`
	Data          string         `gorm:"type:text;default:''"` // JSON: position, nodeCategory, nodeKind, description 等
	RequestPath   string         `gorm:"default:''"`           // 请求地址（不包含 base_url）
	RequestMethod string         `gorm:"default:''"`           // 请求方法，如 GET/POST
	RequestData   string         `gorm:"type:text;default:''"` // 请求体/参数（如 JSON）
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime:nano"`
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

// SessionInfo 会话主表：关联状态机，当前状态与运行状态
type SessionInfo struct {
	ID               int64          `gorm:"primaryKey"`
	SMID             int64          `gorm:"not null;index:idx_sm_logical,unique"`
	LogicalSessionID int64          `gorm:"not null;default:0;index:idx_sm_logical,unique"` // 逻辑会话 id，0 表示设计页会话
	State            string         `gorm:"not null;default:''"`                             // 当前所处状态（节点/状态名）
	Status           string         `gorm:"not null;default:running"`                        // running | ended | suspended
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	CreatedAt        time.Time      `gorm:"autoCreateTime:nano"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime:nano"`
}

// SessionDetail 会话历史明细：一次状态迁移记录（按 session+node 存在则更新）
type SessionDetail struct {
	ID           int64          `gorm:"primaryKey"`
	SessionID    int64          `gorm:"not null;index:idx_session_node,unique"`
	NodeID       string         `gorm:"not null;size:128;index:idx_session_node,unique"` // 节点 id，与 session 确定唯一
	SMID         int64          `gorm:"not null"`
	Event        string         `gorm:"not null;default:''"`
	FromState    string         `gorm:"not null;default:''"`
	ToState      string         `gorm:"not null;default:''"`
	Path         string         `gorm:"default:''"`            // 请求路径
	RequestData  string         `gorm:"type:text;default:''"`  // 请求数据（如 JSON）
	ResponseData string         `gorm:"type:text;default:''"`  // 响应数据（如 JSON）
	Input        string         `gorm:"default:''"`
	Output       string         `gorm:"default:''"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	CreatedAt    time.Time      `gorm:"autoCreateTime:nano"`
}

func Migrate() {
	tables := []any{
		&SMFlow{},
		&SMNode{},
		&SMEdge{},
		&SMApiKey{},
		&SessionInfo{},
		&SessionDetail{},
	}
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log.Fatal(err)
		}
	}
}
