package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type ConsistencySeverity string

const (
	ConsistencyBlocker ConsistencySeverity = "blocker"
	ConsistencyWarning ConsistencySeverity = "warning"
)

type MaterialIssue struct {
	Code     string              `json:"code"`
	Field    string              `json:"field,omitempty"`
	Severity ConsistencySeverity `json:"severity"`
	Message  string              `json:"message"`
}

type MaterialConsistencyReport struct {
	DraftID           string          `json:"draft_id"`
	DraftVersion      int64           `json:"draft_version"`
	TemplateVersionID string          `json:"template_version_id"`
	Scenario          TargetScenario  `json:"scenario"`
	Issues            []MaterialIssue `json:"issues"`
	VisibleFields     []string        `json:"visible_fields"`
	ExportFields      []string        `json:"export_fields"`
	PrivateOmissions  []string        `json:"private_omissions"`
	UnmappedFields    []string        `json:"unmapped_fields"`
	PreviewReady      bool            `json:"preview_ready"`
	ExportReady       bool            `json:"export_ready"`
	ConsistencyDigest string          `json:"consistency_digest"`
}

func EvaluateMaterialConsistency(template TemplateVersion, draft Draft, privacy PrivacyPolicy, scenario TargetScenario) MaterialConsistencyReport {
	report := MaterialConsistencyReport{
		DraftID: draft.ID, DraftVersion: draft.Version, TemplateVersionID: template.ID, Scenario: scenario,
		Issues: make([]MaterialIssue, 0), VisibleFields: make([]string, 0), ExportFields: make([]string, 0),
		PrivateOmissions: make([]string, 0), UnmappedFields: make([]string, 0),
	}
	if draft.TemplateVersionID != template.ID {
		report.Issues = append(report.Issues, MaterialIssue{Code: "TEMPLATE_VERSION_MISMATCH", Severity: ConsistencyBlocker, Message: "草稿绑定的模板版本与检查目标不一致"})
	}
	if !template.Supports(scenario) {
		report.Issues = append(report.Issues, MaterialIssue{Code: "SCENARIO_NOT_SUPPORTED", Severity: ConsistencyBlocker, Message: "模板不适用于当前求职目标"})
	}
	if template.Status == TemplateDeprecated {
		report.Issues = append(report.Issues, MaterialIssue{Code: "TEMPLATE_DEPRECATED", Severity: ConsistencyWarning, Message: "当前草稿仍可使用，但建议切换到在售模板版本"})
	} else if template.Status != TemplatePublished {
		report.Issues = append(report.Issues, MaterialIssue{Code: "TEMPLATE_NOT_PUBLISHED", Severity: ConsistencyBlocker, Message: "未发布模板不能生成正式求职材料"})
	}

	known := make(map[string]struct{})
	digestValues := make(map[string]any)
	for _, field := range template.Fields() {
		known[field.Key] = struct{}{}
		if !field.Visibility.Visible(scenario, draft.Values) {
			continue
		}
		report.VisibleFields = append(report.VisibleFields, field.Key)
		value, present := draft.Values[field.Key]
		if !present || isEmptyValue(value) {
			if field.Required {
				report.Issues = append(report.Issues, MaterialIssue{Code: "REQUIRED_FIELD_MISSING", Field: field.Key, Severity: ConsistencyBlocker, Message: "可见必填字段尚未填写"})
			}
			continue
		}
		digestValues[field.Key] = value
		if field.Private && !privacy.Includes(field.Key) {
			report.PrivateOmissions = append(report.PrivateOmissions, field.Key)
			continue
		}
		report.ExportFields = append(report.ExportFields, field.Key)
	}
	for field := range draft.Values {
		if _, exists := known[field]; !exists {
			report.UnmappedFields = append(report.UnmappedFields, field)
		}
	}
	for field := range draft.Unmapped {
		if _, exists := known[field]; !exists {
			report.UnmappedFields = append(report.UnmappedFields, field)
		}
	}
	report.UnmappedFields = uniqueSorted(report.UnmappedFields)
	for _, field := range report.UnmappedFields {
		report.Issues = append(report.Issues, MaterialIssue{Code: "UNMAPPED_FIELD_PRESERVED", Field: field, Severity: ConsistencyWarning, Message: "字段已保留，但不会出现在当前模板中"})
	}
	sort.Strings(report.VisibleFields)
	sort.Strings(report.ExportFields)
	sort.Strings(report.PrivateOmissions)
	report.PreviewReady = !report.hasBlocker()
	report.ExportReady = report.PreviewReady && draft.Status == DraftActive
	if draft.Status != DraftActive {
		report.Issues = append(report.Issues, MaterialIssue{Code: "DRAFT_ARCHIVED", Severity: ConsistencyBlocker, Message: "归档草稿不能创建新导出"})
	}
	report.ConsistencyDigest = consistencyDigest(template.ID, draft.Version, scenario, digestValues, report.ExportFields)
	return report
}

func (r MaterialConsistencyReport) hasBlocker() bool {
	for _, issue := range r.Issues {
		if issue.Severity == ConsistencyBlocker {
			return true
		}
	}
	return false
}

func consistencyDigest(templateID string, draftVersion int64, scenario TargetScenario, values map[string]any, exportFields []string) string {
	payload, _ := json.Marshal(struct {
		TemplateID   string         `json:"template_id"`
		DraftVersion int64          `json:"draft_version"`
		Scenario     TargetScenario `json:"scenario"`
		Values       map[string]any `json:"values"`
		ExportFields []string       `json:"export_fields"`
	}{templateID, draftVersion, scenario, values, exportFields})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
