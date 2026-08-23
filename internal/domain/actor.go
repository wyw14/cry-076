package domain

type Role string

const (
	RoleOwner    Role = "owner"
	RoleReviewer Role = "reviewer"
	RoleAdmin    Role = "admin"
)

type Actor struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}

func (a Actor) CanManage(ownerID string) bool {
	return a.Role == RoleAdmin || (a.Role == RoleOwner && a.ID == ownerID)
}

func (a Actor) CanReview() bool {
	return a.Role == RoleReviewer || a.Role == RoleAdmin
}
