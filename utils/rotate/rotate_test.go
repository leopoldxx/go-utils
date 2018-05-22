package rotate

import "testing"

func TestRotate(t *testing.T) {

	testCases := []struct {
		raw    string
		sep    string
		expect string
	}{
		{
			raw:    "ab.cd.e.fg",
			sep:    ".",
			expect: "fg.e.cd.ab",
		},
		{
			raw:    "a.b.c.d.e.f.g",
			sep:    "...",
			expect: "a.b.c.d.e.f.g",
		},
		{
			raw:    "ab.cd.ef.g",
			sep:    "",
			expect: "g.fe.dc.ba",
		},
	}

	for _, tc := range testCases {
		if Rotate(tc.raw, tc.sep) != tc.expect {
			t.Fatalf("rotate failed: %v", tc)
		}
	}

}
