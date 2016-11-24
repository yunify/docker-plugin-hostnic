package driver

import (
	"fmt"
	"testing"
)

func TestFindNic(t *testing.T) {
	d := New()
	d.findNics()
	for k, nic := range d.allnic {
		println(k, fmt.Sprintf("%+v", nic))
	}
}
