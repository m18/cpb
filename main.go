package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/db"
	"github.com/m18/cpb/protos"
)

// TODO: add commands at root, e.g., config to print config

func main() {
	cfg, err := config.New(os.Args[1:], os.DirFS)
	exitIf(err)

	p, err := protos.New(cfg.Proto.C, cfg.Proto.Dir, os.DirFS, nil, false)
	exitIf(err)

	db, err := db.New(cfg.DB, p, cfg.InMessages, cfg.OutMessages)
	exitIf(err)

	ctx := context.Background()
	err = db.Ping(ctx)
	exitIf(err)

	cols, rows, err := db.Query(ctx, cfg.DB.Query)
	exitIf(err)

	fmt.Println(cols, rows)
}

func exitIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
