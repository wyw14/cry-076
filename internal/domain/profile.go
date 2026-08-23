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

type ProfileWriteStage string

const (
	ProfileWriteReceived ProfileWriteStage = "received"
	ProfileWriteReplied  ProfileWriteStage = "replied"
	ProfileWriteFailed   ProfileWriteStage = "failed"
)

type ProfileWriteReceipt struct {
	stage       ProfileWriteStage
	candidate   Profile
	baseVersion int64
}

func NewProfileWriteReceipt(candidate Profile, baseVersion int64) ProfileWriteReceipt {
	return ProfileWriteReceipt{
		stage:       ProfileWriteReceived,
		candidate:   copyCandidateProfile(candidate),
		baseVersion: baseVersion,
	}
}

func (r ProfileWriteReceipt) RepositoryAccepted(_ Profile) ProfileWriteReceipt {
	if r.stage != ProfileWriteReceived {
		return r
	}
	r.stage = ProfileWriteReplied
	return r
}

func (r ProfileWriteReceipt) RepositoryFailed() ProfileWriteReceipt {
	if r.stage == ProfileWriteReceived {
		r.stage = ProfileWriteFailed
	}
	return r
}

func (r ProfileWriteReceipt) ClientProfile() Profile {
	if r.stage != ProfileWriteReplied {
		return Profile{}
	}
	response := copyCandidateProfile(r.candidate)
	response.Version = r.baseVersion
	return response
}

func (r ProfileWriteReceipt) Stage() ProfileWriteStage { return r.stage }

func copyCandidateProfile(source Profile) Profile {
	copy := source
	copy.Experiences = append([]Experience(nil), source.Experiences...)
	copy.Skills = append([]Skill(nil), source.Skills...)
	copy.Projects = append([]PortfolioProject(nil), source.Projects...)
	copy.AttachmentIDs = append([]string(nil), source.AttachmentIDs...)
	for index := range copy.Experiences {
		copy.Experiences[index].Highlights = append([]string(nil), source.Experiences[index].Highlights...)
	}
	for index := range copy.Projects {
		copy.Projects[index].Technologies = append([]string(nil), source.Projects[index].Technologies...)
	}
	return copy
}
