package domain

import "time"

type Attachment struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	MediaType string    `json:"media_type"`
	Size      int64     `json:"size"`
	SHA256    string    `json:"sha256"`
	Path      string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

func (a Attachment) Validate(maxSize int64, allowed map[string]bool) error {
	if a.OwnerID == "" || a.Name == "" || a.Size <= 0 || a.Size > maxSize || !allowed[a.MediaType] {
		return ErrValidation
	}
	return nil
}
