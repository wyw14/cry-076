package domain

import "time"

type ContactProfile struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Summary  string `json:"summary"`
}

type Experience struct {
	ID           string     `json:"id"`
	Organization string     `json:"organization"`
	Role         string     `json:"role"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	Highlights   []string   `json:"highlights"`
}

type Skill struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type PortfolioProject struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url,omitempty"`
}

type Profile struct {
	OwnerID       string             `json:"owner_id"`
	Contact       ContactProfile     `json:"contact"`
	Experiences   []Experience       `json:"experiences"`
	Skills        []Skill            `json:"skills"`
	Projects      []PortfolioProject `json:"projects"`
	AttachmentIDs []string           `json:"attachment_ids"`
	Version       int64              `json:"version"`
}
