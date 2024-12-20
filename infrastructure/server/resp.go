package server

import (
	"log"

	"github.com/alifpay/providers.io/pkg"
	"github.com/valyala/fasthttp"
)

func response(ctx *fasthttp.RequestCtx, status int, body any) {
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType("application/json")
	data, err := pkg.Marshal(body)
	if err != nil {
		// todo send to sentry
		log.Println("server response", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.Response.SetBody(data)
}
