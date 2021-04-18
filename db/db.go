package db

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/protos"

	_ "github.com/lib/pq"
)

type DB struct {
	c *sql.DB
	p *queryParser
}

func New(cfg *config.DBConfig, protos *protos.Protos, inMessages map[string]*config.InMessage, outMessages map[string]*config.OutMessage) (*DB, error) {
	connStr, err := connStrGens[cfg.Driver](cfg)
	if err != nil {
		return nil, err
	}
	c, err := sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, err
	}

	return &DB{
		c: c,
		p: newQueryParser(cfg.Driver, protos, inMessages, outMessages),
	}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	c := make(chan error, 1)
	go func() { c <- d.c.PingContext(ctx) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (d *DB) Close() error {
	return d.c.Close()
}

func (d *DB) Query(ctx context.Context, q string) (cols []string, rows [][]interface{}, err error) {
	q, inMessageArgs, outMessageStringers, err := d.p.parse(q)
	if err != nil {
		return nil, nil, err
	}

	rws, err := d.query(ctx, q, btoi(inMessageArgs)...)
	if err != nil {
		return nil, nil, err
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
	rows, err = createRows(rws, colNames, colValTpls, outMessageStringers)
	return cols, rows, err
}

func (d *DB) query(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	resc := make(chan *sql.Rows, 1)
	errc := make(chan error, 1)
	go func() {
		res, err := d.c.QueryContext(ctx, q, args...)
		if err != nil {
			errc <- err
		} else {
			resc <- res
		}
	}()
	select {
	case err := <-errc:
		return nil, err
	case res := <-resc:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func getColData(ct []*sql.ColumnType) ([]string, []interface{}) {
	colNames := make([]string, 0, len(ct))
	colValTpls := make([]interface{}, 0, len(ct))
	for _, c := range ct {
		// fmt.Printf("Name: %s, type: %v\n", c.Name(), c.ScanType())
		colNames = append(colNames, c.Name())
		// although it would be very easy to grab just .Interface() instead of .Elem.Interface()
		// AND pass the resulting slice of pointers-interafces to rows.Scan(slice...),
		// there would be no automatic way of getting the value from such an interface{} (an interface that wraps a pointer to some value)
		// - this would require a switch on type/type assertion and pointer dereference to get the actual value so that,
		// for example, fmt.Sprint(iface) returns the string value of iface, not its memory address, e.g., "5" instead of "0xc000124020", fmt.Sprint(iface) instead of fmt.Sprint(*iface.(*int))
		// instead, this approach uses slice of value-interfaces which we can turn into a temp slice of pointer-interfaces,
		// pass it to rows.Scan(slice...), and then simply see the results in the original slice with no need for a type switch
		// and specialized per type logic
		colValTpls = append(colValTpls, reflect.New(c.ScanType()).Elem().Interface())
	}
	return colNames, colValTpls
}

func createRows(rows *sql.Rows, colNames []string, colValTpls []interface{}, outMessageStringers map[string]func([]byte) (string, error)) ([][]interface{}, error) {
	colValTplPtrs := make([]interface{}, 0, len(colValTpls))
	// a range loop won't work here because `for _, x := range colValTpls` would _copy_ the value into `x`
	// and `&x` would not be pointing to the original value
	for i := 0; i < len(colValTpls); i++ {
		colValTplPtrs = append(colValTplPtrs, &colValTpls[i])
	}

	res := make([][]interface{}, 0, len(colNames))
	for rows.Next() {
		resi := make([]interface{}, 0, len(colNames))
		rows.Scan(colValTplPtrs...)
		for i, dbVal := range colValTpls {
			v, err := getValue(dbVal, outMessageStringers[colNames[i]])
			if err != nil {
				return nil, err
			}
			resi = append(resi, v)
		}
		res = append(res, resi)
	}
	// rows.Close() has been called implicitly as the result of rows.Next() returning false
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func btoi(params [][]byte) []interface{} {
	if params == nil {
		return nil
	}
	res := make([]interface{}, 0, len(params))
	for _, p := range params {
		res = append(res, p)
	}
	return res
}

func getValue(dbVal interface{}, outMessageStringer func([]byte) (string, error)) (interface{}, error) {
	if dbVal == nil || outMessageStringer == nil {
		return dbVal, nil
	}
	b, ok := dbVal.([]byte)
	if !ok {
		return dbVal, nil
	}
	return outMessageStringer(b)
}
