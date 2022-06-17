package main

import (
	"github.com/go-zoox/fs"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	app.Static("/assets", fs.CurrentDir())

	v1 := app.Group("/v1")
	{
		v1.Get("/", func(ctx *zoox.Context) {
			ctx.Write([]byte("v1"))
		})
		v1.Get("/hello", func(ctx *zoox.Context) {
			ctx.JSON(200, zoox.H{
				"hello": "world",
			})
		})
	}

	p := proxy.NewSingleTarget("http://127.0.0.1:8001", &proxy.SingleTargetConfig{})

	app.Fallback(zoox.WrapH(p))

	app.Run()
}
