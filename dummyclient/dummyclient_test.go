package dummyclient

import (
	"fmt"
	"testing"
)

func TestNextNsTime(t *testing.T) {
	avg := int64(40)
	v := int64(0)
	for i := 0; i < 20; i++ {
		f := getNextNsTime(avg)
		v += f
	}
	fmt.Printf("Testing nextNsTime totalled %d\n", v)
}
