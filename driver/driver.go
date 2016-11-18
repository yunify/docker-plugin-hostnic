package driver

import (
	"fmt"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/yunify/docker-plugin-hostnic/log"
	"net"
	"strings"
)

const (
	networkType      = "hostnic"
	excludeNicOption = "exclude_nic"
)

var defaultExcludeNic = [2]string{"eth0", "ens3"}

//HostNicDriver implements github.com/docker/go-plugins-helpers/network.Driver
type HostNicDriver struct {
	network     string
	excludeNics []string
	endpoints   map[string]string
	allNics     []string
}

func (d *HostNicDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return &network.CapabilitiesResponse{Scope: network.LocalScope}, nil
}

func (d *HostNicDriver) CreateNetwork(r *network.CreateNetworkRequest) error {
	log.Debug("CreateNetwork Called: [ %+v ]", r)
	if d.network != "" {
		fmt.Errorf("only one instance of %s network is allowed,  network [%s] exist.", networkType, d.network)
	}
	d.network = r.NetworkID
	d.excludeNics = getExcludeNic(r)

	return nil
}
func (d *HostNicDriver) AllocateNetwork(r *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.Debug("AllocateNetwork Called: [ %+v ]", r)
	return nil, nil
}
func (d *HostNicDriver) DeleteNetwork(r *network.DeleteNetworkRequest) error {
	log.Debug("DeleteNetwork Called: [ %+v ]", r)
	d.network = ""
	return nil
}
func (d *HostNicDriver) FreeNetwork(r *network.FreeNetworkRequest) error {
	log.Debug("FreeNetwork Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) CreateEndpoint(r *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.Debug("CreateEndpoint Called: [ %+v ]", r)
	log.Debug("r.Interface: [ %+v ]", r.Interface)
	return nil, nil
}
func (d *HostNicDriver) DeleteEndpoint(r *network.DeleteEndpointRequest) error {
	log.Debug("DeleteEndpoint Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) EndpointInfo(r *network.InfoRequest) (*network.InfoResponse, error) {
	log.Debug("EndpointInfo Called: [ %+v ]", r)
	//TODO
	res := &network.InfoResponse{
		Value: make(map[string]string),
	}
	return res, nil
}
func (d *HostNicDriver) Join(r *network.JoinRequest) (*network.JoinResponse, error) {
	log.Debug("Join Called: [ %+v ]", r)
	if sID, ok := d.endpoints[r.EndpointID]; ok {
		return nil, fmt.Errorf("Endpoint %s has bean bind to %s", r.EndpointID, sID)
	}
	nic, err := net.InterfaceByName(r.EndpointID)
	if err != nil {
		return nil, err
	}
	d.endpoints[r.EndpointID] = r.SandboxKey
	resp := network.JoinResponse{
		InterfaceName:         network.InterfaceName{SrcName: nic.Name, DstPrefix: "eth"},
		DisableGatewayService: true,
	}
	return &resp, nil
}
func (d *HostNicDriver) Leave(r *network.LeaveRequest) error {
	log.Debug("Join Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) DiscoverNew(r *network.DiscoveryNotification) error {
	log.Debug("DiscoverNew Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) DiscoverDelete(r *network.DiscoveryNotification) error {
	log.Debug("DiscoverDelete Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) ProgramExternalConnectivity(r *network.ProgramExternalConnectivityRequest) error {
	log.Debug("ProgramExternalConnectivity Called: [ %+v ]", r)
	return nil
}
func (d *HostNicDriver) RevokeExternalConnectivity(r *network.RevokeExternalConnectivityRequest) error {
	log.Debug("RevokeExternalConnectivity Called: [ %+v ]", r)
	return nil
}

func getExcludeNic(r *network.CreateNetworkRequest) []string {
	var excludeNics []string
	if r.Options != nil {
		if value, ok := r.Options[excludeNicOption].(string); ok {
			excludeNics = strings.Split(value, ".")
		}
	} else {
		copy(excludeNics[:], defaultExcludeNic[:])
	}
	return excludeNics
}

func getNics() []net.Interface {
	nics, err := net.Interfaces()
	result := make([]net.Interface, 0, len(nics))
	if err == nil {
		for _, nic := range nics {
			//exclude lo
			if nic.Name != "lo" {
				result = append(result, nic)
			}
		}
	} else {

	}
	return result
}
