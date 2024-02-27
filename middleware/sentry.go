package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-zoox/zoox"
)

// reference:
//		sentryhttp  - code: https://github.com/getsentry/sentry-go/blob/master/http/sentryhttp.go
//		echo 			  - code: https://github.com/getsentry/sentry-go/blob/master/echo/sentryecho.go
//		gin 				- code: https://github.com/getsentry/sentry-go/blob/master/gin/sentrygin.go

// SentryOption ...
type SentryOption struct {
	// Repanic configures whether Sentry should repanic after recovery, in most cases it should be set to true,
	// as zoox includes it's own Recover middleware what handles http responses.
	Repanic bool
	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// Because Zoox's Recover handler doesn't restart the application,
	// it's safe to either skip this option or set it to false.
	WaitForDelivery bool
	// Timeout for the event delivery requests.
	Timeout time.Duration
}

// The identifier of the Zoox SDK.
const sdkIdentifier = "sentry.go.zoox"
const valuesKey = "sentry"

var IsSentryInitialized = false

// Sentry ...
func Sentry(opts ...func(opt *SentryOption)) zoox.Middleware {
	if !IsSentryInitialized {
		panic("sentry: Sentry has not been initialized yet, " +
			"should be initialized on the top of application with " +
			"`middleware.InitSentry(middleware.InitSentryOption{ Dsn: '' })`")
	}

	opt := &SentryOption{
		Timeout: 2 * time.Second,
		Repanic: true,
	}
	for _, o := range opts {
		o(opt)
	}

	recoverWithSentry := func(hub *sentry.Hub, r *http.Request) {
		if err := recover(); err != nil {
			eventID := hub.RecoverWithContext(
				context.WithValue(r.Context(), sentry.RequestContextKey, r),
				err,
			)
			if eventID != nil && opt.WaitForDelivery {
				hub.Flush(opt.Timeout)
			}

			if opt.Repanic {
				panic(err)
			}
		}
	}

	return func(ctx *zoox.Context) {
		hub := sentry.GetHubFromContext(ctx.Request.Context())
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
		}

		if client := hub.Client(); client != nil {
			client.SetSDKIdentifier(sdkIdentifier)
		}

		hub.Scope().SetRequest(ctx.Request)
		ctx.State().Set(valuesKey, hub)
		defer recoverWithSentry(hub, ctx.Request)

		ctx.Next()
	}
}

type InitSentryOption = sentry.ClientOptions

// InitSentry ...
func InitSentry(opt InitSentryOption) {
	if IsSentryInitialized {
		panic("sentry: Sentry has been initialized already, should not be initialized more than once.")
	}
	IsSentryInitialized = true

	if opt.Dsn == "" {
		panic("sentry: DSN is required for initializing Sentry")
	}

	err := sentry.Init(opt)
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

// FinishSentry ...
func FinishSentry() {
	sentry.Flush(time.Second)
}

// GetHubFromContext retrieves attached *sentry.Hub instance from echo.Context.
func GetHubFromContext(ctx *zoox.Context) *sentry.Hub {
	if hub, ok := ctx.State().Get(valuesKey).(*sentry.Hub); ok {
		return hub
	}

	return nil
}
