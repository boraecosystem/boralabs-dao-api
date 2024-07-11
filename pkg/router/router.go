package router

import (
	v1 "boralabs/pkg/router/rest/v1"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var E *gin.Engine

func init() {
	E = gin.New()
	E.RemoveExtraSlash = true
	E.Use(gin.Recovery())
	E.Use(gin.Logger())
	E.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// custom format
		return fmt.Sprintf("%s - [%s - %s] \"%s %s %d %d %s \"%s\" \"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339Nano),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	// CORS Middleware
	E.Use(CORSMiddleware)

	// health check
	E.GET("hc", func(c *gin.Context) {
		c.Status(http.StatusOK)
		return
	})

	// REST API V1 Routes
	v1Group := E.Group("")
	rest := v1.REST{}
	rest.RoutesV1(v1Group)
}

func CORSMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("origin")) // any origin
	c.Header("Access-Control-Allow-Methods", http.MethodGet)

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
