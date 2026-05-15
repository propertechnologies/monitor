package gin

import "testing"

func TestStatusClass(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{100, "1xx"},
		{199, "1xx"},
		{200, "2xx"},
		{204, "2xx"},
		{301, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{500, "5xx"},
		{503, "5xx"},
		{0, "unknown"},
		{-1, "unknown"},
	}
	for _, c := range cases {
		got := statusClass(c.code)
		if got != c.want {
			t.Errorf("statusClass(%d) = %q, want %q", c.code, got, c.want)
		}
	}
}
