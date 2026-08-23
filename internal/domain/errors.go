package domain

type RuleCode string

const (
	CodeNotFound    RuleCode = "not_found"
	CodeForbidden   RuleCode = "forbidden"
	CodeConflict    RuleCode = "version_conflict"
	CodeTransition  RuleCode = "invalid_transition"
	CodeIdempotency RuleCode = "idempotency_reuse"
	CodeValidation  RuleCode = "validation"
)

type RuleError struct {
	Code    RuleCode
	Message string
}

func (e *RuleError) Error() string { return e.Message }
func (e *RuleError) Is(target error) bool {
	other, ok := target.(*RuleError)
	return ok && e.Code == other.Code
}

var (
	ErrNotFound          = &RuleError{Code: CodeNotFound, Message: "resume resource not found"}
	ErrForbidden         = &RuleError{Code: CodeForbidden, Message: "resume operation forbidden"}
	ErrConflict          = &RuleError{Code: CodeConflict, Message: "resume version conflict"}
	ErrInvalidTransition = &RuleError{Code: CodeTransition, Message: "resume state transition rejected"}
	ErrIdempotencyReuse  = &RuleError{Code: CodeIdempotency, Message: "resume idempotency key reused"}
	ErrValidation        = &RuleError{Code: CodeValidation, Message: "resume validation failed"}
)
