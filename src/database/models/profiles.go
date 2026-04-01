package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ProfileStatus controls visibility / moderation of a profile.
type ProfileStatus string

const (
	ProfileStatusActive    ProfileStatus = "active"
	ProfileStatusHidden    ProfileStatus = "hidden"
	ProfileStatusSuspended ProfileStatus = "suspended"
)

// MediaKind categorizes stored objects (avatar vs gallery vs documents).
type MediaKind string

const (
	MediaKindAvatar   MediaKind = "avatar"
	MediaKindCover    MediaKind = "cover"
	MediaKindGallery  MediaKind = "gallery"
	MediaKindDocument MediaKind = "document"
)

// Profile is the public-facing user record. Heavy blobs live in object storage;
// we store keys / optional CDN URLs on Profile and ProfileMedia.
type Profile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	DisplayName string `gorm:"size:255"`
	Bio         string `gorm:"type:text"`
	Headline    string `gorm:"size:500"`
	Locale      string `gorm:"size:35"` // BCP-47, e.g. en-GB
	Timezone    string `gorm:"size:64"` // IANA, e.g. Europe/Stockholm
	Phone       string `gorm:"size:40"`

	// Denormalized primary avatar object key (e.g. S3 key); use Media rows for galleries / versions.
	AvatarStorageKey string `gorm:"size:1024"`

	// Links: JSON array of { "label": "...", "url": "https://..." }
	Links datatypes.JSON `gorm:"type:jsonb"`

	// Preferences: arbitrary JSON (notifications, theme, etc.)
	Preferences datatypes.JSON `gorm:"type:jsonb"`

	Status ProfileStatus `gorm:"type:varchar(32);not null;default:active"`

	Media []ProfileMedia `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}

func (Profile) TableName() string {
	return "profiles"
}
