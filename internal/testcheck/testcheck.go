package testcheck

import "testing"

func FatalIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func FatalIfUnexpected(t *testing.T, err error, expectErr bool) {
	if err == nil == expectErr {
		var pre, post string
		if expectErr {
			pre = "not "
		} else {
			post = " not"
		}
		t.Fatalf("expected err %sto be nil but it was%s: %v", pre, post, err)
	}
}
