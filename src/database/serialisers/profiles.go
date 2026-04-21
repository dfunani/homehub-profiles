package serialisers

import (
	"dfunani/homehub-profiles/src/database/models"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProfileSerialiser is the API shape for a profile (no nested user model).
type ProfileSerialiser struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	UserID           uuid.UUID           `json:"user_id"`
	DisplayName      string              `json:"display_name"`
	Bio              string              `json:"bio"`
	Headline         string              `json:"headline"`
	Locale           string              `json:"locale"`
	Timezone         string              `json:"timezone"`
	Phone            string              `json:"phone"`
	AvatarStorageKey string              `json:"avatar_storage_key"`
	Links            []map[string]string `json:"links"`       // raw JSON array; stored as jsonb
	Preferences      map[string]string   `json:"preferences"` // raw JSON object
	Status           string              `json:"status"`

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
	var links []map[string]string
	err := json.Unmarshal(m.Links, &links)
	if err != nil {
		panic("Failed to marshal links: " + err.Error())
	}
	out.Links = links

	var preferences map[string]string
	err = json.Unmarshal(m.Preferences, &preferences)
	if err != nil {
		panic("Failed to marshal preferences: " + err.Error())
	}
	out.Preferences = preferences

	for i := range m.Media {
		item := (&ProfileMediaSerialiser{}).FromModel(&m.Media[i])
		if item != nil {
			out.Media = append(out.Media, *item)
		}
	}
	return out
}

func (p *ProfileSerialiser) ToModel() *models.Profile {
	m := &models.Profile{
		UserID:           p.UserID,
		DisplayName:      p.DisplayName,
		Bio:              p.Bio,
		Headline:         p.Headline,
		Locale:           p.Locale,
		Timezone:         p.Timezone,
		Phone:            p.Phone,
		AvatarStorageKey: p.AvatarStorageKey,
	}
	if len(p.Links) > 0 {
		links, err := json.Marshal(p.Links)
		if err != nil {
			panic("Failed to marshal links: " + err.Error())
		}
		m.Links = datatypes.JSON(links)
	}
	if len(p.Preferences) > 0 {
		preferences, err := json.Marshal(p.Preferences)
		if err != nil {
			panic("Failed to marshal preferences: " + err.Error())
		}
		m.Preferences = datatypes.JSON(preferences)
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

func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// CreateProfile persists a new profile. IDs and timestamps are filled by GORM/DB when empty.
func CreateProfile(db *gorm.DB, p *ProfileSerialiser) *ProfileSerialiser {
	m := p.ToModel()
	if err := db.Create(m).Error; err != nil {
		log.Printf("[Profile] Error creating profile: %s", err.Error())
		if isDuplicateKeyError(err) {
			log.Printf("[Profile] Profile already exists for user: %s", m.UserID.String())
			panic("Profile already exists for user: " + m.UserID.String())
		}
		panic("Failed to create profile: " + err.Error())
	}
	return p.FromModel(m)
}

// GetProfile loads a profile by id (optionally preloads media).
func GetProfile(db *gorm.DB, id uuid.UUID, preloadMedia bool) *ProfileSerialiser {
	q := db
	if preloadMedia {
		q = q.Preload("Media", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at ASC")
		})
	}
	var m models.Profile
	if err := q.First(&m, "id = ?", id).Error; err != nil {
		log.Printf("[Profile] Error getting profile: %s", err.Error())
		panic("Failed to get profile: " + err.Error())
	}
	return (&ProfileSerialiser{}).FromModel(&m)
}

// GetProfileByUserID loads the profile for a user id.
func GetProfileByUserID(db *gorm.DB, userID uuid.UUID, preloadMedia bool) *ProfileSerialiser {
	q := db
	if preloadMedia {
		q = q.Preload("Media", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at ASC")
		})
	}
	var m models.Profile
	if err := q.Where("user_id = ?", userID).First(&m).Error; err != nil {
		log.Printf("[Profile] Error getting profile by user_id: %s", err.Error())
		panic("Failed to get profile by user_id: " + err.Error())
	}
	return (&ProfileSerialiser{}).FromModel(&m)
}
