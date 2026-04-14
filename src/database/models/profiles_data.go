package models

import (
	"time"

	"github.com/google/uuid"
)

// MediaKind categorizes stored objects (avatar vs gallery vs documents).
type MediaKind string

const (
	MediaKindAvatar   MediaKind = "avatar"
	MediaKindCover    MediaKind = "cover"
	MediaKindGallery  MediaKind = "gallery"
	MediaKindDocument MediaKind = "document"
	MediaKindOther    MediaKind = "other"
)

// ProfileMedia stores references to images or files in external storage.
type ProfileMedia struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null;gen_random_uuid()"`

	StorageKey string `gorm:"size:1024;not null"`
	// Optional stable public URL if you front S3 with CloudFront and cache the string.
	PublicURL   string    `gorm:"size:2048"`
	Kind        MediaKind `gorm:"type:varchar(32);not null;default:other"`
	Caption     string    `gorm:"size:500"`
	Width       *int
	Height      *int
	ContentType string `gorm:"size:128"`

	Profile *Profile `gorm:"foreignKey:ProfileID"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (ProfileMedia) TableName() string {
	return "profile_media"
}
