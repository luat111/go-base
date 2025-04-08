package middlewares

import (
	"context"
	"go-base/pkg/common"

	"github.com/gin-gonic/gin"
)

func AttachParams(c *gin.Context) {
	params := c.Params
	m := make(map[string]string, len(params))

	for _, v := range params {
		m[v.Key] = v.Value
	}

	ctx := context.WithValue(c.Request.Context(), common.ReqParams, m)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}
