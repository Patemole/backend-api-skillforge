package models

import (
	"github.com/google/uuid"
)

type Job struct {
	ID        int64          `json:"id,omitempty"`
	Type      string         `json:"type"`
	UserID    uuid.UUID      `json:"user_id"`
	Payload   map[string]any `json:"payload"`
	Status    string         `json:"status"`
	Result    map[string]any `json:"result,omitempty"`
	Error     *string        `json:"error,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}
