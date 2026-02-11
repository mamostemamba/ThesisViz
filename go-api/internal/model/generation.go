package model

import (
	"time"

	"github.com/google/uuid"
)

type Generation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	ParentID     *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	Format       string     `gorm:"type:varchar(20);not null" json:"format"`
	Prompt       string     `gorm:"type:text;not null" json:"prompt"`
	Status       string     `gorm:"type:varchar(20);default:'queued'" json:"status"`
	Code         *string    `gorm:"type:text" json:"code,omitempty"`
	ImageKey     *string    `gorm:"type:varchar(255)" json:"image_key,omitempty"`
	Explanation  *string    `gorm:"type:text" json:"explanation,omitempty"`
	ReviewIssues *string    `gorm:"type:jsonb" json:"review_issues,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
	Parent  *Generation `gorm:"foreignKey:ParentID" json:"-"`
}
