package main

import (
	"context"
	"fmt"
)

func main() {
	cfg, err := newConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)

	res, err := cfg.InMessages["p"].JSON([]string{"1", "\"Pops\""})
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.InMessages["p"].params, res)

	// testRun(cfg)
	// testPB()
}

func testPB() {
	p, err := newProtos("test/proto")
	if err != nil {
		panic(err)
	}
	bytes, err := p.protoBytes("tutorial.Person", `{"name": "bound", "phones":[{"number": "1-23", "type": "HOME", "is_main": true}]}`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Worked!", bytes)
}

func testRun(cfg *config) {
	p, err := newProtos("test/proto")
	if err != nil {
		panic(err)
	}
	db, err := newDB(cfg.Driver, cfg.ConnStr, cfg.InMessages, p)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := db.ping(ctx); err != nil {
		panic(err)
	}
	cols, rows, err := db.query(ctx, "select * from samples")
	if err != nil {
		panic(err)
	}
	fmt.Println(cols, rows)
}
