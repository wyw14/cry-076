package middleware

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		c.Next()
	}
}

func CORS(allowed []string) gin.HandlerFunc {
	policy := compileOriginPolicy(allowed)
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Vary", "Origin")
		}
		if policy.allows(origin) {
			writeCORSHeaders(c, origin)
		}
		if c.Request.Method == "OPTIONS" {
			if origin != "" && !policy.allows(origin) {
				c.AbortWithStatus(403)
				return
			}
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

type originPolicy struct {
	exact map[string]struct{}
}

func compileOriginPolicy(allowed []string) originPolicy {
	policy := originPolicy{exact: make(map[string]struct{}, len(allowed))}
	for _, candidate := range allowed {
		normalized, ok := normalizeOrigin(candidate)
		if !ok {
			continue
		}
		policy.exact[normalized] = struct{}{}
	}
	return policy
}

func (p originPolicy) allows(candidate string) bool {
	normalized, valid := normalizeOrigin(candidate)
	if !valid {
		return false
	}
	if len(p.exact) == 0 {
		return true
	}
	_, allowed := p.exact[normalized]
	return allowed
}

func normalizeOrigin(candidate string) (string, bool) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || candidate == "null" {
		return "", false
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}
	if parsed.Host == "" || parsed.Path != "" {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", false
	}
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}
	return parsed.Scheme + "://" + host + ":" + port, true
}

func writeCORSHeaders(c *gin.Context, origin string) {
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "false")
	c.Header("Access-Control-Allow-Headers", strings.Join([]string{
		"Content-Type", "X-Actor-ID", "X-Actor-Role", "X-Request-ID",
		"Idempotency-Key", "If-Match",
	}, ", "))
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Max-Age", "600")
}
