package utils

import "testing"

func testStatusDot(t *testing.T) {
	var tt = []struct {
		Color    string
		Expected string
	}{
		{"red", "@k\u2B24 @k\u2B24 @r\u2B24  @{|}(red)"},
		{"yellow", "@k\u2B24 @y\u2B24 @k\u2B24  @{|}(yellow)"},
		{"green", "@g\u2B24 @k\u2B24 @k\u2B24  @{|}(green)"},
		{"purple", "@k\u2B24 @k\u2B24 @k\u2B24  @{|}(none)"},
	}
	for _, test := range tt {
		dots := StatusDots(test.Color)
		if dots != test.Expected {
			t.Errorf("%s expected, got: %s", test.Expected, dots)
		}
	}

}
