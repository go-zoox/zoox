package zoox

func HealthCheck() HandlerFunc {
	return func(ctx *Context) {
		ctx.String(200, "OK")
	}
}
