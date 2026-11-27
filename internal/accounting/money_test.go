package accounting

import "testing"

func TestMicroDollarPrecision(t *testing.T) {
	for _, test := range []struct{ input, output, inputRate, outputRate, want int64 }{{1, 0, 499999, 0, 0}, {1, 0, 500000, 0, 1}, {1, 1, 500000, 500000, 1}, {0, 0, 1000000, 1000000, 0}, {1000000, 1000000, 1, 3, 4}} {
		actual, err := Charge(test.input, test.output, test.inputRate, test.outputRate)
		if err != nil || actual != test.want {
			t.Fatal("micro precision", actual, err)
		}
	}
	if _, err := Charge(-1, 0, 0, 0); err == nil {
		t.Fatal("negative accepted")
	}
	if _, err := Charge(9223372036854775807, 0, 9223372036854775807, 0); err == nil {
		t.Fatal("overflow accepted")
	}
}
