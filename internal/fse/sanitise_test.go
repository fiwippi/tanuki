package fse

import "testing"

func TestSanitise(t *testing.T) {
	a := "it's"
	if a != Sanitise(a) {
		t.Fail()
	}

	a = "!!!"
	if a != Sanitise(a) {
		t.Fail()
	}

	a = "it??s"
	if a == Sanitise(a) {
		t.Fail()
	}
}
