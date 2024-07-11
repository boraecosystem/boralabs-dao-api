package main

import (
	_ "boralabs/config"
	"boralabs/internal/event_logger"
	"boralabs/pkg/datastore/mongodb"
	"boralabs/pkg/notification"
	"boralabs/pkg/router"
	"boralabs/pkg/util"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func init() {
	mongodb.New()
	notification.BaseLoggers = append(notification.BaseLoggers, &notification.SlackLogger)
	go eventCollect()
}

// RateLimiter Define RateLimiter struct
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter Initialize RateLimiter with NewRateLimiter function
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

// Middleware Implement Rate Limiting middleware with Middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func eventCollect() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// init
	event_logger.Collector{}.Collect()
	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println(err)
						util.PrintStackTrace()
					}
				}()
				event_logger.Collector{}.Collect()
			}()
		}
	}
}

func main() {
	defer func() {
		if err := mongodb.Conn.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Initialize Rate Limiter
	rateLimiter := NewRateLimiter(5, 10)

	// Apply Rate Limiting middleware
	router.E.Use(rateLimiter.Middleware())

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", port()),
		Handler:        router.E, // set router
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
		IdleTimeout:    120 * time.Second,
	}
	http.DefaultClient.Timeout = 10 * time.Second // set default http timeout

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func port() string {
	if os.Getenv("PORT") != "" {
		return os.Getenv("PORT")
	}
	return "8080"
}
