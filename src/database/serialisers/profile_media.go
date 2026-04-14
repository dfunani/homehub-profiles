package serialisers

import (
	"dfunani/homehub-profiles/src/database/models"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileMediaSerialiser is the API shape for profile_media (images / files in storage).
type ProfileMediaSerialiser struct {
	ID        uuid.UUID `json:"id,omitempty"`
	ProfileID uuid.UUID `json:"profile_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	StorageKey  string `json:"storage_key"`
	PublicURL   string `json:"public_url,omitempty"`
	Kind        string `json:"kind"`
	Caption     string `json:"caption,omitempty"`
	Width       *int   `json:"width,omitempty"`
	Height      *int   `json:"height,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

func (m *ProfileMediaSerialiser) FromModel(x *models.ProfileMedia) *ProfileMediaSerialiser {
	if x == nil {
		return nil
	}
	return &ProfileMediaSerialiser{
		ID:          x.ID,
		ProfileID:   x.ProfileID,
		CreatedAt:   x.CreatedAt,
		UpdatedAt:   x.UpdatedAt,
		StorageKey:  x.StorageKey,
		PublicURL:   x.PublicURL,
		Kind:        string(x.Kind),
		Caption:     x.Caption,
		Width:       x.Width,
		Height:      x.Height,
		ContentType: x.ContentType,
	}
}

func (m *ProfileMediaSerialiser) ToModel() *models.ProfileMedia {
	kind := models.MediaKindGallery
	if m.Kind != "" {
		kind = models.MediaKind(m.Kind)
	}
	return &models.ProfileMedia{
		ID:          m.ID,
		ProfileID:   m.ProfileID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		StorageKey:  m.StorageKey,
		PublicURL:   m.PublicURL,
		Kind:        kind,
		Caption:     m.Caption,
		Width:       m.Width,
		Height:      m.Height,
		ContentType: m.ContentType,
	}
}

func (m *ProfileMediaSerialiser) ToJSON() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

func (m *ProfileMediaSerialiser) FromJSON(data []byte) *ProfileMediaSerialiser {
	if err := json.Unmarshal(data, m); err != nil {
		return nil
	}
	return m
}

// CreateProfileMedia inserts a media row for a profile.
func CreateProfileMedia(db *gorm.DB, m *ProfileMediaSerialiser) *ProfileMediaSerialiser {
	x := m.ToModel()
	if err := db.Create(x).Error; err != nil {
		panic("Failed to create profile media: " + err.Error())
	}
	return m.FromModel(x)
}

// GetProfileMedia loads one media row by id.
func GetProfileMedia(db *gorm.DB, id uuid.UUID) *ProfileMediaSerialiser {
	var x models.ProfileMedia
	if err := db.First(&x, "id = ?", id).Error; err != nil {
		panic("Failed to get profile media: " + err.Error())
	}
	return (&ProfileMediaSerialiser{}).FromModel(&x)
}
