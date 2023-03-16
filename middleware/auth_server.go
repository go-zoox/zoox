package middleware

import (
	"fmt"

	"github.com/go-zoox/fetch"
	"github.com/go-zoox/zoox"
)

// AuthServerConfig ...
type AuthServerConfig struct {
	Server string `json:"server"`
}

// AuthServer is a middleware that authenticates via Auth Server.
func AuthServer(cfg *AuthServerConfig) zoox.Middleware {
	if cfg.Server == "" {
		panic("server is required")
	}

	return func(ctx *zoox.Context) {
		// 1. Bear Token
		if token, ok := ctx.BearerToken(); ok {
			if status, code, message, err := handleAuthServerTypeBearerToken(ctx, cfg.Server, token); err != nil {
				ctx.Logger.Errorf("[auth-server: bearer token] failed to authenticate with auth server: %s", err)

				ctx.JSON(status, zoox.H{
					"code":    code,
					"message": message,
				})
				return
			}

			ctx.Next()
			return
		}

		// 2. Basic Auth
		if username, password, ok := ctx.Request.BasicAuth(); ok {
			if _, _, _, err := handleAuthServerTypeBasicAuth(ctx, cfg.Server, username, password); err != nil {
				ctx.Logger.Errorf("[auth-server: bearer token] failed to authenticate with auth server: %s", err)

				ctx.Set("WWW-Authenticate", `Basic realm="Go-Zoox"`)
				ctx.Status(401)
				return
			}

			ctx.Next()
			return
		}

		ctx.JSON(401, zoox.H{
			"code":    4001000,
			"message": "only support bear token or basic auth",
		})
	}
}

func handleAuthServerTypeBearerToken(ctx *zoox.Context, server string, token string) (status int, code int, message string, err error) {
	var response *fetch.Response
	typ := "bearertoken"

	response, err = fetch.Post(server, &fetch.Config{
		Headers: fetch.Headers{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		Body: map[string]any{
			"type":  typ,
			"token": token,
		},
	})
	if err != nil {
		err = fmt.Errorf("[auth-server: %s] failed to connect to auth server: %s", typ, err)
		status = 500
		code = 500101
		message = fmt.Sprintf("[auth-server: %s] failed to connect to auth server", typ)
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("[auth-server: %s] auth server response error: %s", typ, response.String())
		status = 500
		code = 500102
		message = fmt.Sprintf("[auth-server: %s] auth server response error", typ)
		return
	}

	return
}

func handleAuthServerTypeBasicAuth(ctx *zoox.Context, server string, username, password string) (status int, code int, message string, err error) {
	var response *fetch.Response
	typ := "basicauth"

	response, err = fetch.Post(server, &fetch.Config{
		Headers: fetch.Headers{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		Body: map[string]any{
			"type":     typ,
			"username": username,
			"password": password,
		},
	})
	if err != nil {
		err = fmt.Errorf("[auth-server: %s] failed to connect to auth server: %s", typ, err)
		status = 500
		code = 500103
		message = fmt.Sprintf("[auth-server: %s] failed to connect to auth server", typ)
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("[auth-server: %s] auth server response error: %s", typ, response.String())
		status = 500
		code = 500104
		message = fmt.Sprintf("[auth-server: %s] auth server response error", typ)
		return
	}

	return
}
