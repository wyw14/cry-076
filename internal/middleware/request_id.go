package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

type CorrelationStage string

const (
	CorrelationAccepted  CorrelationStage = "accepted"
	CorrelationAllocated CorrelationStage = "allocated"
	CorrelationAttached  CorrelationStage = "attached"
)

type CorrelationFlow struct {
	id       string
	stage    CorrelationStage
	external bool
}

func AcceptCorrelation(value string) CorrelationFlow {
	id := strings.TrimSpace(value)
	if id == "" {
		return CorrelationFlow{stage: CorrelationAccepted}
	}
	return CorrelationFlow{id: id, stage: CorrelationAccepted, external: true}
}

func (f CorrelationFlow) Allocate() CorrelationFlow {
	if f.id != "" {
		return f
	}
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err == nil {
		f.id = hex.EncodeToString(raw[:])
	} else {
		f.id = "unknown"
	}
	f.stage = CorrelationAllocated
	return f
}

func (f CorrelationFlow) Attach() CorrelationFlow {
	if f.id == "" {
		f = f.Allocate()
	}
	f.stage = CorrelationAttached
	return f
}

func (f CorrelationFlow) Value() string           { return f.id }
func (f CorrelationFlow) Stage() CorrelationStage { return f.stage }
func (f CorrelationFlow) FromCaller() bool        { return f.external }

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		flow := AcceptCorrelation(c.GetHeader("X-Request-ID")).Attach()
		c.Set(RequestIDKey, flow)
		c.Header("X-Request-ID", flow.Value())
		c.Next()
	}
}
func GetRequestID(c *gin.Context) string {
	value, _ := c.Get(RequestIDKey)
	id, _ := value.(string)
	return id
}
