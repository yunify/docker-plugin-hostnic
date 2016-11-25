package driver

import (
	"fmt"
	"testing"
)

func TestFindNic(t *testing.T) {
	d := New()
	d.findNic(func(nic *HostNic) bool {
		println(fmt.Sprintf("Find nic [%+v] ", nic))
		println("Addr", nic.Addr())
		println("HardwareAddr", nic.HardwareAddr())
		return false
	})
}

func TestFindNicByAddr(t *testing.T) {
	addr := "127.0.0.1/8"
	d := New()
	nic := d.FindNicByAddr(addr)
	if nic == nil {
		t.Fatal("can not find nic by addr: ", addr)
	}
}
