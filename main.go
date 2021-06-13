package main

import (
	"bufio"
	"context"
	"os"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/db"
	"github.com/m18/cpb/printer"
	"github.com/m18/cpb/protos"
	"github.com/m18/cpb/sys"
)

// TODO: add commands at root, e.g., config to print config

func main() {
	cfg, err := config.New(os.Args[1:], os.DirFS)
	sys.ExitIf(err)

	args, pipe, err := querySources(cfg)
	sys.ExitIf(err)
	if !args && !pipe {
		return
	}

	p, err := protos.New(cfg.Proto.C, cfg.Proto.Dir, cfg.Proto.Deterministic, os.DirFS, nil, false)
	sys.ExitIf(err)

	db, err := db.New(cfg.DB, p, cfg.InMessages, cfg.OutMessages, cfg.AutoMapOutMessages)
	sys.ExitIf(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sys.HandleInterrupt(ctx, cancel)

	err = db.Ping(ctx)
	sys.ExitIf(err, ctx.Err())

	// TODO: config
	pr, err := printer.New(
		os.Stdout,
		printer.WithFormat(printer.FormatTable),
		printer.WithHeader(true),
		printer.WithSpacing(1),
	)
	sys.ExitIf(err)

	if args {
		err = queryAndPrint(ctx, db, cfg.DB.Query, pr)
		sys.ExitIf(err, ctx.Err())
	}

	if pipe {
		err = queryFromPipe(ctx, db, pr)
		sys.ExitIf(err, ctx.Err())
	}
}

func querySources(cfg *config.Config) (args, pipe bool, err error) {
	args = cfg.DB.Query != ""
	pipe, err = sys.IsPipedIn()
	return args, pipe, err
}

func queryFromPipe(ctx context.Context, db *db.DB, pr *printer.Printer) error {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		if err := queryAndPrint(ctx, db, s.Text(), pr); err != nil {
			return err
		}
	}
	return nil
}

func queryAndPrint(ctx context.Context, db *db.DB, q string, pr *printer.Printer) error {
	if q == "" {
		return nil
	}
	cols, rows, err := db.Query(ctx, q)
	if err != nil {
		return err
	}
	pr.Print(cols, rows)
	return nil
}
