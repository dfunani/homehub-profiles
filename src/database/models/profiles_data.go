package models

import (
	"time"

	"github.com/google/uuid"
)

// ProfileMedia stores references to images or files in external storage.
type ProfileMedia struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	StorageKey string `gorm:"size:1024;not null"`
	// Optional stable public URL if you front S3 with CloudFront and cache the string.
	PublicURL   string    `gorm:"size:2048"`
	Kind        MediaKind `gorm:"type:varchar(32);not null"`
	Caption     string    `gorm:"size:500"`
	SortOrder   int64     `gorm:"not null;default:0"`
	Width       *int
	Height      *int
	ContentType string `gorm:"size:128"`

	Profile *Profile `gorm:"foreignKey:ProfileID"`
}

func (ProfileMedia) TableName() string {
	return "profile_media"
}
