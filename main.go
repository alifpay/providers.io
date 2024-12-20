package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/alifpay/providers.io/infrastructure/server"
	"github.com/alifpay/providers.io/pkg"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	db, err := pkg.ConnectDB(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pkg.CloseDB(db)
	g, ctx := errgroup.WithContext(ctx)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	server.Start(ctx, g)

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Println("server exited with error: ", err)
	}
}
