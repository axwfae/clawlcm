package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID        int64
	Version   int
	Name      string
	AppliedAt time.Time
}

func RunMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&ConversationModel{},
		&MessageModel{},
		&SummaryModel{},
		&ContextItemModel{},
		&Migration{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	if !db.Migrator().HasColumn(&Migration{}, "applied_at") {
		if err := db.Migrator().RenameColumn(&Migration{}, "AppliedAt", "applied_at"); err != nil {
			fmt.Printf("migration column rename warning: %v\n", err)
		}
	}

	return nil
}

type ConversationModel struct {
	ID           int64  `gorm:"primaryKey"`
	SessionKey   string `gorm:"uniqueIndex;size:512"`
	SessionID    string `gorm:"size:128"`
	MessageCount int    `gorm:"default:0"`
	TokenCount   int    `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ConversationModel) TableName() string {
	return "conversations"
}

type MessageModel struct {
	ID             int64  `gorm:"primaryKey"`
	ConversationID int64  `gorm:"index;not null"`
	Ordinal        int    `gorm:"not null"`
	Role           string `gorm:"size:32;not null"`
	Content        string `gorm:"type:text"`
	TokenCount     int    `gorm:"default:0"`
	CreatedAt      time.Time
}

func (MessageModel) TableName() string {
	return "messages"
}

type SummaryModel struct {
	ID             int64  `gorm:"primaryKey"`
	ConversationID int64  `gorm:"index;not null"`
	SummaryType    string `gorm:"size:32;not null"`
	Depth          int    `gorm:"default:0"`
	Content        string `gorm:"type:text"`
	TokenCount     int    `gorm:"default:0"`
	SourceTokens   int    `gorm:"default:0"`
	Ordinal        int    `gorm:"not null"`
	ParentIDs      string `gorm:"type:text"`
	SourceIDs      string `gorm:"type:text"`
	CreatedAt      time.Time
}

func (SummaryModel) TableName() string {
	return "summaries"
}

type ContextItemModel struct {
	ID             int64  `gorm:"primaryKey"`
	ConversationID int64  `gorm:"index;not null"`
	ItemType       string `gorm:"size:32;not null"`
	ItemID         int64  `gorm:"not null"`
	Ordinal        int    `gorm:"not null"`
	TokenCount     int    `gorm:"default:0"`
	Keywords       string `gorm:"type:text"`
	CreatedAt      time.Time
}

func (ContextItemModel) TableName() string {
	return "context_items"
}
