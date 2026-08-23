package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-076/internal/middleware"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, allowedOrigins []string, ready func() bool, handlers Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		middleware.CORS(allowedOrigins),
		middleware.ResumeWorkflowLogger(logger),
		middleware.ResumeWorkspaceRecovery(logger),
	)
	registerServiceChecks(router, ready)
	registerResumeWorkspace(router.Group("/api/v1", middleware.Actor()), handlers)
	return router
}

func registerServiceChecks(router *gin.Engine, ready func() bool) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "resume-consistency"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		if !ready() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "resume_catalog"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "dependency": "resume_catalog"})
	})
}

func registerResumeWorkspace(api *gin.RouterGroup, handlers Handlers) {
	registerTemplateGovernance(api, handlers)
	registerDraftContinuity(api, handlers)
	registerPrivateMaterials(api, handlers)
	registerLocalDelivery(api, handlers)
}

func registerTemplateGovernance(api *gin.RouterGroup, handlers Handlers) {
	api.GET("/templates", handlers.listTemplates)
	api.POST("/templates", handlers.createTemplate)
	api.POST("/templates/:id/transitions", handlers.transitionTemplate)
	api.GET("/feedback", handlers.listFeedback)
	api.POST("/feedback", handlers.createFeedback)
	api.POST("/feedback/:id/transitions", handlers.reviewFeedback)
}

func registerDraftContinuity(api *gin.RouterGroup, handlers Handlers) {
	api.GET("/drafts", handlers.listDrafts)
	api.POST("/drafts", handlers.createDraft)
	api.GET("/drafts/:id", handlers.getDraft)
	api.GET("/drafts/:id/consistency", handlers.inspectDraftConsistency)
	api.PUT("/drafts/:id/autosave", handlers.saveDraft)
	api.DELETE("/drafts/:id", handlers.archiveDraft)
	api.POST("/drafts/:id/snapshots", handlers.createSnapshot)
	api.GET("/drafts/:id/snapshots", handlers.listSnapshots)
	api.POST("/drafts/:id/switch-template", handlers.switchTemplate)
	api.POST("/snapshots/:snapshot_id/restore", handlers.restoreSnapshot)
	api.GET("/snapshots/compare", handlers.compareSnapshots)
}

func registerPrivateMaterials(api *gin.RouterGroup, handlers Handlers) {
	api.GET("/profiles/:owner_id", handlers.getProfile)
	api.PUT("/profiles/:owner_id", handlers.saveProfile)
	api.GET("/privacy/:owner_id", handlers.getPrivacy)
	api.PUT("/privacy/:owner_id", handlers.savePrivacy)
	api.POST("/attachments/:owner_id", handlers.uploadAttachment)
	api.GET("/attachments/:id/content", handlers.openAttachment)
}

func registerLocalDelivery(api *gin.RouterGroup, handlers Handlers) {
	api.POST("/exports", handlers.createExport)
	api.POST("/exports/:id/validate", handlers.validateExport)
	api.GET("/audits", handlers.listAudits)
}
