package driver

import (
	"fmt"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/hashicorp/netlink"
	"os"
	"path"
	"testing"
)

func TestLink(t *testing.T) {
	links, err := netlink.LinkList()
	if err != nil {
		print(err)
		t.Fatal(err)
	}
	for _, link := range links {
		println(fmt.Sprintf("%+v", link))
	}
}

func TestConfig(t *testing.T) {
	os.Remove(path.Join(configDir, "config.json"))

	driver, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	ipv4data := &network.IPAMData{
		Gateway:      "192.168.0.1/24",
		Pool:         "192.168.0.0/24",
		AddressSpace: "LocalDefault",
	}
	err = driver.RegisterNetwork("0", ipv4data)
	if err != nil {
		t.Fatal(err)
	}

	ipv4data1 := &network.IPAMData{
		Gateway:      "192.168.1.1/24",
		Pool:         "192.168.1.0/24",
		AddressSpace: "LocalDefault",
	}
	err = driver.RegisterNetwork("1", ipv4data1)
	if err != nil {
		t.Fatal(err)
	}

	driver.saveConfig()

	driver2, _ := New()

	if len(driver2.networks) != 2 {
		t.Fatal("expect networks len is 2")
	}
}
