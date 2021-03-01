package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"text/template"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TODO: rename test to example, move init into data

func main() {
	if ok, pwd, err := isEscape(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if ok {
		fmt.Println(url.QueryEscape(pwd))
		return
	}

	cfg, err := newConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// fmt.Println(len(cfg.InMessages))
	// fmt.Println(cfg.OutMessages["p"].template.Execute(os.Stdout, map[string]interface{}{"name": "blah", "address_postcode": 2010}))

	// res, err := cfg.InMessages["p"].JSON([]string{"1", "\"Pops\""})
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(cfg.InMessages["p"].params, res)

	testRun(cfg)
	// testPB()
	// testTpl()
}

func testPB() {
	p, err := newProtos("test/proto")
	if err != nil {
		panic(err)
	}
	bytes, err := p.protoBytes("tutorial.Person", `{
		"name": "bound", 
		"phones":[
			{"number": "1-23", "type": "HOME", "is_main": true}
		], 
		"home_address":{"postcode": 2010},
		"pairs": {"one": 1},
		"onestr": "one-one"
	}`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Worked!", bytes)

	///////////////////////

	d, err := p.fileReg.FindDescriptorByName("tutorial.Person")
	if err != nil {
		panic(err)
	}
	d2, err := p.fileReg.FindDescriptorByName("tutorial.Person.PhoneNumber")
	if err != nil {
		panic(err)
	}
	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		panic("not a message descriptor")
	}
	md2, ok := d2.(protoreflect.MessageDescriptor)
	if !ok {
		panic("not a message descriptor")
	}
	mt := dynamicpb.NewMessageType(md)
	rm := mt.New()
	m := rm.Interface()
	if err := proto.Unmarshal(bytes, m); err != nil {
		panic(err)
	}
	fd := md.Fields().ByName("phones")
	if fd == nil {
		panic("invalid field name")
	}
	fd2 := md2.Fields().ByName("number")
	if fd2 == nil {
		panic("invalid field name")
	}
	fv := rm.Get(fd).Interface()
	fmt.Println("Worked too!", fd.IsList(), rm.Get(fd).List().Get(0).Message().Get(fd2), fv, "--->", m)

	rm.Range(recRange)
}

// TODO: handle maps and lists
func recRange(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	fmt.Println(fd.FullName(), fd.Kind(), fd.Kind() == protoreflect.MessageKind, fd.IsList(), v.Interface())
	if (fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind) && !fd.IsList() && !fd.IsMap() {
		v.Message().Range(recRange)
	}
	return true
}

func testTpl() {
	if tpl, err := template.New("").Parse("blah: {{.val}}"); err != nil {
		panic(err)
	} else {
		tpl.Execute(os.Stdout, map[string]string{"val": "yo"})
	}
}

func testRun(cfg *config) {
	p, err := newProtos("test/proto")
	if err != nil {
		panic(err)
	}
	db, err := newDB(cfg.DB, p, cfg.InMessages, cfg.OutMessages)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := db.ping(ctx); err != nil {
		panic(err)
	}
	fmt.Println("------->", cfg.DB.Query)
	cols, rows, err := db.query(ctx, cfg.DB.Query /*"select * from samples"*/)
	if err != nil {
		panic(err)
	}
	fmt.Println(cols, rows)
}
