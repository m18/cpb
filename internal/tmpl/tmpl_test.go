package tmpl

import "testing"

func TestPropToTemplateParam(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "", expected: ""},
		{input: "foo", expected: "foo"},
		{input: "foo.bar", expected: "foo_bar"},
		{input: "...", expected: "___"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			res := PropToTemplateParam(test.input)
			if res != test.expected {
				t.Errorf("expected %q but got %q", test.expected, res)
			}
		})
	}
}
