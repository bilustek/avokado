package avokadodb

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel represents core base model for avokado application.
type BaseModel struct {
	ID        uint      `json:"-"          gorm:"primarykey"`
	UID       uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

// DBModelizer defines model identifier for generic storage usage.
type DBModelizer interface {
	TableName() string
	GetPublicID() uuid.UUID
}
