package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-076/internal/domain"
)

const ActorKey = "actor"

func Actor() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Actor-ID")
		role := domain.Role(c.GetHeader("X-Actor-Role"))
		if id == "" || !validRole(role) {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "UNAUTHENTICATED", "message": "缺少有效身份", "request_id": GetRequestID(c)}})
			return
		}
		c.Set(ActorKey, domain.Actor{ID: id, Role: role})
		c.Next()
	}
}
func GetActor(c *gin.Context) domain.Actor {
	value, _ := c.Get(ActorKey)
	actor, _ := value.(domain.Actor)
	return actor
}
func validRole(role domain.Role) bool {
	return role == domain.RoleOwner || role == domain.RoleReviewer || role == domain.RoleAdmin
}
