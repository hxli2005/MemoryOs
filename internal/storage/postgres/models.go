package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/datatypes"
)

// DialogueMemoryPO 对话记忆持久化对象
// 对应数据库表: dialogue_memory
type DialogueMemoryPO struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID       string           `gorm:"type:varchar(255);not null;index:idx_dialogue_user_created"`
	SessionID    *string          `gorm:"type:varchar(255);index:idx_dialogue_session"`
	Content      string           `gorm:"type:text;not null"`
	Role         string           `gorm:"type:varchar(50);default:user"`
	Embedding    *pgvector.Vector `gorm:"type:vector(768)"`
	Metadata     datatypes.JSON   `gorm:"type:jsonb;default:'{}'"`
	MemoryType   string           `gorm:"type:varchar(50);not null;default:user_message;index:idx_dialogue_type"`
	Importance   float64          `gorm:"type:double precision;default:0.6"`
	AccessCount  int              `gorm:"type:integer;default:0"`
	LastAccessed *time.Time       `gorm:"type:timestamptz"`
	CreatedAt    time.Time        `gorm:"type:timestamptz;default:now();index:idx_dialogue_user_created"`
	UpdatedAt    time.Time        `gorm:"type:timestamptz;default:now()"`
}

// TableName 指定表名
func (DialogueMemoryPO) TableName() string {
	return "dialogue_memory"
}

// TopicMemoryPO 话题记忆持久化对象
// 对应数据库表: topic_memory
type TopicMemoryPO struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID       string           `gorm:"type:varchar(255);not null;index:idx_topic_user_created"`
	Title        string           `gorm:"type:varchar(500);not null"`
	Summary      *string          `gorm:"type:text"`
	Content      string           `gorm:"type:text;not null"`
	Embedding    *pgvector.Vector `gorm:"type:vector(768)"`
	Keywords     datatypes.JSON   `gorm:"type:text[]"` // PostgreSQL 数组类型
	DialogueIDs  datatypes.JSON   `gorm:"type:bigint[]"`
	Metadata     datatypes.JSON   `gorm:"type:jsonb;default:'{}'"`
	MemoryType   string           `gorm:"type:varchar(50);not null;default:topic_thread"`
	Importance   float64          `gorm:"type:double precision;default:0.7"`
	AccessCount  int              `gorm:"type:integer;default:0"`
	LastAccessed *time.Time       `gorm:"type:timestamptz"`
	CreatedAt    time.Time        `gorm:"type:timestamptz;default:now();index:idx_topic_user_created"`
	UpdatedAt    time.Time        `gorm:"type:timestamptz;default:now()"`
}

// TableName 指定表名
func (TopicMemoryPO) TableName() string {
	return "topic_memory"
}

// ProfileMemoryPO 用户画像持久化对象
// 对应数据库表: profile_memory
type ProfileMemoryPO struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID       string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_profile_user"`
	Content      string          `gorm:"type:text;not null"` // 画像描述文本
	Preferences  datatypes.JSON  `gorm:"type:jsonb"`
	Habits       datatypes.JSON  `gorm:"type:jsonb"`
	Features     datatypes.JSON  `gorm:"type:jsonb"`
	Embedding    pgvector.Vector `gorm:"type:vector(768)"`
	TopicIDs     datatypes.JSON  `gorm:"type:bigint[]"`
	Metadata     datatypes.JSON  `gorm:"type:jsonb;default:'{}'"`
	MemoryType   string          `gorm:"type:varchar(50);not null;default:user_identity"`
	Importance   float64         `gorm:"type:double precision;default:0.9"`
	AccessCount  int             `gorm:"type:integer;default:0"`
	LastAccessed *time.Time      `gorm:"type:timestamptz"`
	CreatedAt    time.Time       `gorm:"type:timestamptz;default:now()"`
	UpdatedAt    time.Time       `gorm:"type:timestamptz;default:now()"`
}

// TableName 指定表名
func (ProfileMemoryPO) TableName() string {
	return "profile_memory"
}
