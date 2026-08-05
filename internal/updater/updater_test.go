package updater

import "testing"

func TestDiff(t *testing.T) {
	if got := diff([]int{1, 2, 3}, []int{2}); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Errorf("diff = %v, want [1 3]", got)
	}
	if got := diff(nil, []int{1}); got != nil {
		t.Errorf("diff of nil = %v", got)
	}
	if got := diff([]int{1}, nil); len(got) != 1 {
		t.Errorf("diff against nil = %v", got)
	}
}

func TestIntPtrEqual(t *testing.T) {
	one, alsoOne, two := 1, 1, 2
	if !intPtrEqual(nil, nil) {
		t.Error("nil == nil")
	}
	if intPtrEqual(&one, nil) || intPtrEqual(nil, &one) {
		t.Error("nil != value")
	}
	if !intPtrEqual(&one, &alsoOne) {
		t.Error("equal values")
	}
	if intPtrEqual(&one, &two) {
		t.Error("different values")
	}
}
