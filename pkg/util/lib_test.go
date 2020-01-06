package util

import (
	"testing"
)

func TestBtoi(t *testing.T) {
	test := []int{1, 2, 3}
	bytes := Itob(test)
	i := Btoi(bytes)
	if i.([]int)[1] != 2 {
		t.Error("Wrong convertion. Expected 'bazz' got", i)
	}
}
