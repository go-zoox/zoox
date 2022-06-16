package main

import (
	"github.com/go-zoox/fs"
	"github.com/go-zoox/zoox"
)

func main() {
	r := zoox.Default()

	r.Static("/assets", fs.CurrentDir())

	r.Get("/", func(ctx *zoox.Context) {
		ctx.Write([]byte("helloworld"))
	})

	r.Get("/panic", func(ctx *zoox.Context) {
		var a []int
		a[0] = 1
	})

	v1 := r.Group("/v1")
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

	r.Run(":8080")
}
