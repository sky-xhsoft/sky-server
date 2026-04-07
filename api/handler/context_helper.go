package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
)

// withCompanyID 将公司ID添加到context中
func withCompanyID(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if companyID, exists := c.Get("companyID"); exists {
		if id, ok := companyID.(uint); ok {
			ctx = context.WithValue(ctx, contextutil.CompanyIDKey, id)
		}
	} else if companyId, exists := c.Get("companyId"); exists {
		if id, ok := companyId.(uint); ok {
			ctx = context.WithValue(ctx, contextutil.CompanyIDKey, id)
		}
	}
	return ctx
}
