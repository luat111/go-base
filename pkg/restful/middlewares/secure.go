package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

var secureMiddleware = secure.New(secure.Options{
	FrameDeny:        true,
	BrowserXssFilter: true,
})

func SecureMiddleware(c *gin.Context) {
	err := secureMiddleware.Process(c.Writer, c.Request)

	// If there was an error, do not continue.
	if err != nil {
		c.Abort()
		return
	}

	// Avoid header rewrite if response is a redirection.
	if status := c.Writer.Status(); status > 300 && status < 399 {
		c.Abort()
	}
}
