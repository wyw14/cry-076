package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-076/internal/application"
)

func pageRequest(c *gin.Context) application.PageRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return application.PageRequest{Page: page, PageSize: size, Sort: c.DefaultQuery("sort", "updated_at"), Order: c.DefaultQuery("order", "desc")}
}
func expectedVersion(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.GetHeader("If-Match"), 10, 64)
}
