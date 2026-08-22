package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResumeWorkflowLogger records the public material workflow rather than raw
// request bodies, which may contain private profile fields.
func ResumeWorkflowLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		actor := GetActor(c)
		logger.Info("resume_workflow_completed",
			zap.String("request_id", GetRequestID(c)),
			zap.String("workflow", workflowName(c.Request.Method, c.FullPath())),
			zap.String("route", c.FullPath()),
			zap.String("actor_role", string(actor.Role)),
			zap.Int("status_code", c.Writer.Status()),
			zap.Bool("changed_material", c.Request.Method != http.MethodGet),
			zap.Duration("elapsed", time.Since(startedAt)),
		)
	}
}

func ResumeWorkspaceRecovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("resume_workspace_panic",
			zap.String("request_id", GetRequestID(c)),
			zap.String("workflow", workflowName(c.Request.Method, c.FullPath())),
			zap.Any("panic", recovered),
		)
		c.AbortWithStatusJSON(500, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "服务内部错误", "request_id": GetRequestID(c)}})
	})
}

func workflowName(method, route string) string {
	switch {
	case strings.Contains(route, "switch-template"):
		return "template_switch"
	case strings.Contains(route, "snapshots"):
		return "draft_history"
	case strings.Contains(route, "exports"):
		return "local_export"
	case strings.Contains(route, "privacy") || strings.Contains(route, "attachments"):
		return "privacy_and_attachment"
	case strings.Contains(route, "feedback") || strings.Contains(route, "templates"):
		return "template_governance"
	case strings.Contains(route, "drafts") && method != http.MethodGet:
		return "draft_editing"
	case strings.Contains(route, "drafts"):
		return "draft_browsing"
	default:
		return "service_boundary"
	}
}
