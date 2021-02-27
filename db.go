package main

import (
	"context"
	"database/sql"
	"reflect"

	_ "github.com/lib/pq"
)

type db struct {
	c *sql.DB
	// returns query with placeholders, placeholder values, map from colName to pb-pretty-print func, error
	p *queryParser
}

func newDB(driver, connStr string, protos *protos, inMessages map[string]*InMessage, outMessages map[string]*OutMessage) (*db, error) {
	c, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}

	return &db{
		c: c,
		p: newQueryParser(driver, protos, inMessages, outMessages),
	}, nil
}

func (d *db) ping(ctx context.Context) error {
	c := make(chan error, 1)
	go func() { c <- d.c.PingContext(ctx) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (d *db) close() error {
	return d.c.Close()
}

func (d *db) query(ctx context.Context, q string) (cols []string, rows [][]interface{}, err error) {
	query, params, prettyPrinters, err := d.p.Parse(q)
	if err != nil {
		return nil, nil, err
	}

	resc := make(chan *sql.Rows, 1)
	errc := make(chan error, 1)
	go func() {
		rws, err := d.c.QueryContext(ctx, query, d.btoi(params)...)
		if err != nil {
			errc <- err
		} else {
			resc <- rws
		}
	}()
	var rws *sql.Rows
	select {
	case err := <-errc:
		return nil, nil, err
	case rws = <-resc:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	cols, err = rws.Columns()
	if err != nil {
		return nil, nil, err
	}
	colTypes, err := rws.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	colNames, colValTpls := getColData(colTypes)
	rows, err = createRows(rws, colNames, colValTpls, prettyPrinters)
	return cols, rows, err
}

func (d *db) btoi(params [][]byte) []interface{} {
	res := make([]interface{}, 0, len(params))
	for _, p := range params {
		res = append(res, p)
	}
	return res
}

func getColData(ct []*sql.ColumnType) ([]string, []interface{}) {
	colNames := make([]string, 0, len(ct))
	colValTpls := make([]interface{}, 0, len(ct))
	for _, c := range ct {
		// fmt.Printf("Name: %s, type: %v\n", c.Name(), c.ScanType())
		colNames = append(colNames, c.Name())
		// although it would be very easy to grab just .Interface() instead of .Elem.Interface()
		// AND pass the resulting slice of pointers-interafces to rows.Scan(slice...),
		// there would be no automatic way of getting the value from such an interface{} (which is really a *interface{})
		// - this would require a switch on type or something else
		// instead, this approach uses slice of value-interfaces which we can turn into a temp slice of pointer-interfaces,
		// pass it to rows.Scan(slice...), and then simply see the results in the original slice with no need for a type switch
		// and specialized per type logic
		colValTpls = append(colValTpls, reflect.New(c.ScanType()).Elem().Interface())
	}
	return colNames, colValTpls
}

func createRows(rows *sql.Rows, colNames []string, colValTpls []interface{}, prettyPrinters map[string]func([]byte) (string, error)) ([][]interface{}, error) {
	colValTplPtrs := make([]interface{}, 0, len(colValTpls))
	// a range loop won't work here because `for _, x := range colValTpls` would _copy_ a value into `x`
	// and `&x` would not be pointing to the original value
	for i := 0; i < len(colValTpls); i++ {
		colValTplPtrs = append(colValTplPtrs, &colValTpls[i])
	}

	res := make([][]interface{}, 0, len(colNames))
	for rows.Next() {
		resi := make([]interface{}, 0, len(colNames))
		rows.Scan(colValTplPtrs...)
		for i, v := range colValTpls {
			vv, err := getFromValue( /*colNames[i],*/ v, prettyPrinters[colNames[i]])
			if err != nil {
				return nil, err
			}
			resi = append(resi, vv)
		}
		res = append(res, resi)
	}
	return res, nil
}

func getFromValue( /*colName string, */ colVal interface{}, prettyPrinter func([]byte) (string, error)) (interface{}, error) {
	if colVal == nil || prettyPrinter == nil {
		return colVal, nil
	}
	b, ok := colVal.([]byte)
	if !ok {
		return colVal, nil
	}
	return prettyPrinter(b)
}
