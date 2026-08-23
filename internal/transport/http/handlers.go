package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	appmw "github.com/wyw14/cry-076/internal/middleware"
)

type Handlers struct {
	Templates   *application.TemplateService
	Drafts      *application.DraftService
	Switches    *application.SwitchService
	Consistency *application.ConsistencyService
	Profiles    *application.ProfileService
	Privacy     *application.PrivacyService
	Exports     *application.ExportService
	Attachments *application.AttachmentService
	Feedback    *application.FeedbackService
	Audits      *application.AuditService
}

func (h Handlers) listTemplates(c *gin.Context) {
	filter := application.TemplateListFilter{PageRequest: pageRequest(c), Status: domain.TemplateStatus(c.Query("status")), Scenario: domain.TargetScenario(c.Query("scenario")), Category: c.Query("category")}
	items, total, err := h.Templates.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	respondList(c, items, total, filter.Page, filter.PageSize)
}

type createTemplateRequest struct {
	TemplateID string                   `json:"template_id" binding:"required"`
	Version    int64                    `json:"version" binding:"required,min=1"`
	Name       string                   `json:"name" binding:"required"`
	Category   string                   `json:"category" binding:"required"`
	Scenarios  []domain.TargetScenario  `json:"scenarios" binding:"required,min=1"`
	Sections   []domain.TemplateSection `json:"sections" binding:"required,min=1"`
	Style      domain.PreviewStyle      `json:"style" binding:"required"`
}

func (h Handlers) createTemplate(c *gin.Context) {
	var request createTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	item := domain.TemplateVersion{TemplateID: request.TemplateID, Version: request.Version, Name: request.Name, Category: request.Category, Scenarios: request.Scenarios, Sections: request.Sections, Style: request.Style}
	created, err := h.Templates.Create(c.Request.Context(), appmw.GetActor(c), item, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, created)
}

type transitionTemplateRequest struct {
	Status  domain.TemplateStatus `json:"status" binding:"required"`
	Version int64                 `json:"version" binding:"required,min=1"`
}

func (h Handlers) transitionTemplate(c *gin.Context) {
	var request transitionTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	updated, err := h.Templates.Transition(c.Request.Context(), appmw.GetActor(c), c.Param("id"), request.Version, request.Status, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, updated)
}

type createDraftRequest struct {
	OwnerID           string         `json:"owner_id" binding:"required"`
	TargetID          string         `json:"target_id" binding:"required"`
	TemplateVersionID string         `json:"template_version_id" binding:"required"`
	Values            map[string]any `json:"values" binding:"required"`
}

func (h Handlers) createDraft(c *gin.Context) {
	var request createDraftRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	input := application.CreateDraftInput{OwnerID: request.OwnerID, TargetID: request.TargetID, TemplateVersionID: request.TemplateVersionID, Values: request.Values, IdempotencyKey: c.GetHeader("Idempotency-Key"), RequestHash: requestHash(c), RequestID: appmw.GetRequestID(c)}
	created, err := h.Drafts.Create(c.Request.Context(), appmw.GetActor(c), input)
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, created)
}
func (h Handlers) getDraft(c *gin.Context) {
	item, err := h.Drafts.Get(c.Request.Context(), appmw.GetActor(c), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, item)
}
func (h Handlers) inspectDraftConsistency(c *gin.Context) {
	report, err := h.Consistency.Inspect(c.Request.Context(), appmw.GetActor(c), c.Param("id"), domain.TargetScenario(c.Query("scenario")))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, report)
}
func (h Handlers) listDrafts(c *gin.Context) {
	filter := application.DraftListFilter{PageRequest: pageRequest(c), OwnerID: c.Query("owner_id"), Status: domain.DraftStatus(c.Query("status")), TargetID: c.Query("target_id")}
	items, total, err := h.Drafts.List(c.Request.Context(), appmw.GetActor(c), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	respondList(c, items, total, filter.Page, filter.PageSize)
}

type saveDraftRequest struct {
	Values map[string]any `json:"values" binding:"required"`
}

func (h Handlers) saveDraft(c *gin.Context) {
	version, err := expectedVersion(c)
	if err != nil {
		respondError(c, domain.ErrValidation)
		return
	}
	var request saveDraftRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	input := application.SaveDraftInput{DraftID: c.Param("id"), ExpectedVersion: version, Values: request.Values, IdempotencyKey: c.GetHeader("Idempotency-Key"), RequestHash: requestHash(c), RequestID: appmw.GetRequestID(c)}
	updated, err := h.Drafts.AutoSave(c.Request.Context(), appmw.GetActor(c), input)
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, updated)
}

type snapshotRequest struct {
	Reason string `json:"reason" binding:"required,max=120"`
}

func (h Handlers) createSnapshot(c *gin.Context) {
	var request snapshotRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	item, err := h.Drafts.Snapshot(c.Request.Context(), appmw.GetActor(c), c.Param("id"), request.Reason, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, item)
}
func (h Handlers) listSnapshots(c *gin.Context) {
	page := pageRequest(c)
	items, total, err := h.Drafts.ListSnapshots(c.Request.Context(), appmw.GetActor(c), c.Param("id"), page)
	if err != nil {
		respondError(c, err)
		return
	}
	respondList(c, items, total, page.Page, page.PageSize)
}
func (h Handlers) restoreSnapshot(c *gin.Context) {
	version, err := expectedVersion(c)
	if err != nil {
		respondError(c, domain.ErrValidation)
		return
	}
	restored, err := h.Drafts.Restore(c.Request.Context(), appmw.GetActor(c), c.Param("snapshot_id"), version, c.GetHeader("Idempotency-Key"), requestHash(c), appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, restored)
}
func (h Handlers) compareSnapshots(c *gin.Context) {
	comparison, err := h.Drafts.Compare(c.Request.Context(), appmw.GetActor(c), c.Query("left"), c.Query("right"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, comparison)
}
func (h Handlers) archiveDraft(c *gin.Context) {
	version, err := expectedVersion(c)
	if err != nil {
		respondError(c, domain.ErrValidation)
		return
	}
	if err := h.Drafts.Archive(c.Request.Context(), appmw.GetActor(c), c.Param("id"), version); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type switchRequest struct {
	TemplateVersionID string                `json:"template_version_id" binding:"required"`
	Scenario          domain.TargetScenario `json:"scenario" binding:"required"`
	Version           int64                 `json:"version" binding:"required,min=1"`
}

func (h Handlers) switchTemplate(c *gin.Context) {
	var request switchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	result, err := h.Switches.Switch(c.Request.Context(), appmw.GetActor(c), application.SwitchTemplateInput{DraftID: c.Param("id"), TargetTemplateVersionID: request.TemplateVersionID, Scenario: request.Scenario, ExpectedVersion: request.Version, IdempotencyKey: c.GetHeader("Idempotency-Key"), RequestHash: requestHash(c), RequestID: appmw.GetRequestID(c)})
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, result)
}

func (h Handlers) getProfile(c *gin.Context) {
	item, err := h.Profiles.Get(c.Request.Context(), appmw.GetActor(c), c.Param("owner_id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, item)
}
func (h Handlers) saveProfile(c *gin.Context) {
	version, err := expectedVersion(c)
	if err != nil {
		respondError(c, domain.ErrValidation)
		return
	}
	var profile domain.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		respondBindingError(c, err)
		return
	}
	profile.OwnerID = c.Param("owner_id")
	saved, err := h.Profiles.Save(c.Request.Context(), appmw.GetActor(c), profile, version, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, saved)
}

func (h Handlers) getPrivacy(c *gin.Context) {
	item, version, err := h.Privacy.Get(c.Request.Context(), appmw.GetActor(c), c.Param("owner_id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("ETag", strconv.FormatInt(version, 10))
	respondData(c, http.StatusOK, item)
}
func (h Handlers) savePrivacy(c *gin.Context) {
	version, err := expectedVersion(c)
	if err != nil {
		respondError(c, domain.ErrValidation)
		return
	}
	var policy domain.PrivacyPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		respondBindingError(c, err)
		return
	}
	policy.OwnerID = c.Param("owner_id")
	updated, err := h.Privacy.Save(c.Request.Context(), appmw.GetActor(c), policy, version, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("ETag", strconv.FormatInt(updated, 10))
	respondData(c, http.StatusOK, gin.H{"version": updated})
}

type exportRequest struct {
	DraftID    string              `json:"draft_id" binding:"required"`
	SnapshotID string              `json:"snapshot_id"`
	Format     domain.ExportFormat `json:"format" binding:"required"`
}

func (h Handlers) createExport(c *gin.Context) {
	var request exportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	result, err := h.Exports.Create(c.Request.Context(), appmw.GetActor(c), application.CreateExportInput{DraftID: request.DraftID, SnapshotID: request.SnapshotID, Format: request.Format, IdempotencyKey: c.GetHeader("Idempotency-Key"), RequestHash: requestHash(c), RequestID: appmw.GetRequestID(c)})
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, result)
}
func (h Handlers) validateExport(c *gin.Context) {
	result, err := h.Exports.Validate(c.Request.Context(), appmw.GetActor(c), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, result)
}

func (h Handlers) uploadAttachment(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondBindingError(c, err)
		return
	}
	defer file.Close()
	item, err := h.Attachments.Upload(c.Request.Context(), appmw.GetActor(c), c.Param("owner_id"), header.Filename, header.Header.Get("Content-Type"), appmw.GetRequestID(c), file)
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, item)
}
func (h Handlers) openAttachment(c *gin.Context) {
	reader, err := h.Attachments.Open(c.Request.Context(), appmw.GetActor(c), c.Param("id"), appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

type feedbackRequest struct {
	TemplateVersionID string `json:"template_version_id" binding:"required"`
	Kind              string `json:"kind" binding:"required"`
	Message           string `json:"message" binding:"required,max=2000"`
}

func (h Handlers) createFeedback(c *gin.Context) {
	var request feedbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	item, err := h.Feedback.Submit(c.Request.Context(), appmw.GetActor(c), request.TemplateVersionID, request.Kind, request.Message, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusCreated, item)
}

type reviewFeedbackRequest struct {
	Status     domain.FeedbackStatus `json:"status" binding:"required"`
	Resolution string                `json:"resolution"`
}

func (h Handlers) reviewFeedback(c *gin.Context) {
	var request reviewFeedbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondBindingError(c, err)
		return
	}
	item, err := h.Feedback.Review(c.Request.Context(), appmw.GetActor(c), c.Param("id"), request.Status, request.Resolution, appmw.GetRequestID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondData(c, http.StatusOK, item)
}
func (h Handlers) listFeedback(c *gin.Context) {
	filter := application.FeedbackListFilter{PageRequest: pageRequest(c), Status: domain.FeedbackStatus(c.Query("status")), TemplateVersionID: c.Query("template_version_id")}
	items, total, err := h.Feedback.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	respondList(c, items, total, filter.Page, filter.PageSize)
}

func (h Handlers) listAudits(c *gin.Context) {
	filter := application.AuditListFilter{PageRequest: pageRequest(c), ActorID: c.Query("actor_id"), Resource: c.Query("resource"), ResourceID: c.Query("resource_id"), Action: c.Query("action")}
	items, total, err := h.Audits.List(c.Request.Context(), appmw.GetActor(c), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	respondList(c, items, total, filter.Page, filter.PageSize)
}

func requestHash(c *gin.Context) string {
	if value := c.GetHeader("X-Request-Hash"); value != "" {
		return value
	}
	return appmw.GetRequestID(c)
}
