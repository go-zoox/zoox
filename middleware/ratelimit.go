package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-zoox/counter/bucket"
	"github.com/go-zoox/headers"
	"github.com/go-zoox/ratelimit"
	"github.com/go-zoox/zoox"
)

// RateLimitConfig ...
type RateLimitConfig struct {
	Period time.Duration
	Limit  int64
	//
	Namespace string
	//
	RedisHost     string
	RedisPort     int
	RedisDB       int
	RedisPassword string
}

// RateLimit middleware for zoox
func RateLimit(cfg *RateLimitConfig) zoox.Middleware {
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "go-zoox"
	}

	var limiter *ratelimit.RateLimit
	var err error
	if cfg.RedisHost != "" {
		limiter, err = ratelimit.NewRedis(namespace, cfg.Period, cfg.Limit, &bucket.RedisConfig{
			Host:     cfg.RedisHost,
			Port:     cfg.RedisPort,
			DB:       cfg.RedisDB,
			Password: cfg.RedisPassword,
		})
	} else {
		limiter = ratelimit.NewMemory(namespace, cfg.Period, cfg.Limit)
	}

	if err != nil {
		panic(fmt.Errorf("failed to create ratelimit middleware: %s", err))
	}

	return func(ctx *zoox.Context) {
		ip := ctx.Request.RemoteAddr
		limiter.Inc(ip)

		// GitHub Standard
		ctx.Set(headers.XRateLimitRemaining, fmt.Sprintf("%d", limiter.Remaining(ip)))
		ctx.Set(headers.XRateLimitReset, fmt.Sprintf("%d", limiter.ResetAt(ip)/1000))
		ctx.Set(headers.XRateLimitLimit, fmt.Sprintf("%d", limiter.Total(ip)))

		// MDN
		ctx.Set(headers.RetryAfter, fmt.Sprintf("%d", limiter.ResetAfter(ip)))

		if limiter.IsExceeded(ip) {
			ctx.Fail(errors.New("too many requests"), http.StatusTooManyRequests, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		ctx.Next()
	}
}
