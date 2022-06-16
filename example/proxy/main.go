package main

import (
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
)

func main() {
	r := zoox.Default()

	r.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"hello": "world",
		})
	})

	p := proxy.NewSingleTarget("https://httpbin.zcorky.com", &proxy.SingleTargetConfig{
		Rewrites: map[string]string{
			"^/api/(.*)": "/$1",
		},
	})

	r.Get("/api/*path", zoox.WrapH(p))

	r.Run()
}
