package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/savsgio/gotils/strconv"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
)

func Start(ctx context.Context, g *errgroup.Group) {
	addr := ":8080"
	server := &fasthttp.Server{
		Name:                          "Go",
		Handler:                       requestHandler,
		DisableHeaderNamesNormalizing: true,
	}

	g.Go(func() error {
		log.Println("starting server http server at " + addr)
		if err := server.ListenAndServe(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		log.Println("http server shut down gracefully")
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.ShutdownWithContext(shutdownCtx)
		if err != nil {
			return err
		}
		return nil
	})
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	if !isAuth(ctx) {
		return
	}
	switch strconv.B2S(ctx.Request.URI().Path()) {
	case "/pay":
		pay(ctx)
	default:
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound), fasthttp.StatusNotFound)
	}
}

func isAuth(ctx *fasthttp.RequestCtx) bool {
	if auth := strconv.B2S(ctx.Request.Header.Peek("Authorization")); len(auth) == 0 {
		ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
		return false
	} else if auth != "Bearer 123456" {
		ctx.Error("Forbidden", fasthttp.StatusForbidden)
		return false
	}
	return true
}
