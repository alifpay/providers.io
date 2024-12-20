package server

import "github.com/valyala/fasthttp"

func pay(ctx *fasthttp.RequestCtx) {
	test := struct {
		Amount int    `json:"amount"`
		Status string `json:"status"`
	}{
		Amount: 100,
		Status: "success",
	}
	response(ctx, fasthttp.StatusOK, test)
}
