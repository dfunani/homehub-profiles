package serialisers

import (
	"dfunani/homehub-profiles/src/database/models"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProfileSerialiser is the API shape for a profile (no nested user model).
type ProfileSerialiser struct {
	ID        uuid.UUID `json:"id,omitempty"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	DisplayName      string `json:"display_name,omitempty"`
	Bio              string `json:"bio,omitempty"`
	Headline         string `json:"headline,omitempty"`
	Locale           string `json:"locale,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
	Phone            string `json:"phone,omitempty"`
	AvatarStorageKey string `json:"avatar_storage_key,omitempty"`
	Links            []byte `json:"links,omitempty"`        // raw JSON array; stored as jsonb
	Preferences      []byte `json:"preferences,omitempty"`  // raw JSON object
	Status           string `json:"status,omitempty"`

	Media []ProfileMediaSerialiser `json:"media,omitempty"`
}

func (p *ProfileSerialiser) FromModel(m *models.Profile) *ProfileSerialiser {
	if m == nil {
		return nil
	}
	out := &ProfileSerialiser{
		ID:               m.ID,
		UserID:           m.UserID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DisplayName:      m.DisplayName,
		Bio:              m.Bio,
		Headline:         m.Headline,
		Locale:           m.Locale,
		Timezone:         m.Timezone,
		Phone:            m.Phone,
		AvatarStorageKey: m.AvatarStorageKey,
		Status:           string(m.Status),
	}
	if len(m.Links) > 0 {
		out.Links = append([]byte(nil), m.Links...)
	}
	if len(m.Preferences) > 0 {
		out.Preferences = append([]byte(nil), m.Preferences...)
	}
	for i := range m.Media {
		item := (&ProfileMediaSerialiser{}).FromModel(&m.Media[i])
		if item != nil {
			out.Media = append(out.Media, *item)
		}
	}
	return out
}

func (p *ProfileSerialiser) ToModel() *models.Profile {
	st := models.ProfileStatusActive
	if p.Status != "" {
		st = models.ProfileStatus(p.Status)
	}
	m := &models.Profile{
		ID:               p.ID,
		UserID:           p.UserID,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		DisplayName:      p.DisplayName,
		Bio:              p.Bio,
		Headline:         p.Headline,
		Locale:           p.Locale,
		Timezone:         p.Timezone,
		Phone:            p.Phone,
		AvatarStorageKey: p.AvatarStorageKey,
		Status:           st,
	}
	if len(p.Links) > 0 {
		m.Links = datatypes.JSON(append([]byte(nil), p.Links...))
	}
	if len(p.Preferences) > 0 {
		m.Preferences = datatypes.JSON(append([]byte(nil), p.Preferences...))
	}
	return m
}

func (p *ProfileSerialiser) ToJSON() string {
	b, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(b)
}

func (p *ProfileSerialiser) FromJSON(data []byte) *ProfileSerialiser {
	if err := json.Unmarshal(data, p); err != nil {
		return nil
	}
	return p
}

// CreateProfile persists a new profile. IDs and timestamps are filled by GORM/DB when empty.
func CreateProfile(db *gorm.DB, p *ProfileSerialiser) *ProfileSerialiser {
	m := p.ToModel()
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	if m.Status == "" {
		m.Status = models.ProfileStatusActive
	}
	if len(m.Links) == 0 {
		m.Links = datatypes.JSON(`[]`)
	}
	if len(m.Preferences) == 0 {
		m.Preferences = datatypes.JSON(`{}`)
	}
	if err := db.Create(m).Error; err != nil {
		log.Fatalf("Failed to create profile: %v", err)
	}
	return p.FromModel(m)
}

// GetProfile loads a profile by id (optionally preloads media).
func GetProfile(db *gorm.DB, id uuid.UUID, preloadMedia bool) *ProfileSerialiser {
	q := db
	if preloadMedia {
		q = q.Preload("Media", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, created_at ASC")
		})
	}
	var m models.Profile
	if err := q.First(&m, "id = ?", id).Error; err != nil {
		log.Fatalf("Failed to get profile: %v", err)
	}
	return (&ProfileSerialiser{}).FromModel(&m)
}

// GetProfileByUserID loads the profile for a user id.
func GetProfileByUserID(db *gorm.DB, userID uuid.UUID, preloadMedia bool) *ProfileSerialiser {
	q := db
	if preloadMedia {
		q = q.Preload("Media", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, created_at ASC")
		})
	}
	var m models.Profile
	if err := q.Where("user_id = ?", userID).First(&m).Error; err != nil {
		log.Fatalf("Failed to get profile by user_id: %v", err)
	}
	return (&ProfileSerialiser{}).FromModel(&m)
}
