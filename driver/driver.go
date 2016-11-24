package driver

import (
	"errors"
	"fmt"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/docker/libnetwork/netutils"
	"github.com/docker/libnetwork/types"
	"github.com/hashicorp/netlink"
	"github.com/yunify/docker-plugin-hostnic/log"
	"net"
	"strings"
	"sync"
)

const (
	networkType      = "hostnic"
	excludeNicOption = "exclude_nic"

	vethPrefix          = "veth"
	vethLen             = 7
	containerVethPrefix = "eth"
)

var defaultExcludeNic = [2]string{"eth0", "ens3"}

type NicTable map[string]*HostNic

type HostNic struct {
	NetInterface net.Interface
	endpoint     *endpoint
}

func (n *HostNic) HardwareAddr() string {
	return n.NetInterface.HardwareAddr.String()
}

func (n *HostNic) Addr() string {
	addrs, err := n.NetInterface.Addrs()
	if err != nil {
		log.Error("Get interface [%+v] addr error: %s", n.NetInterface, err.Error())
		return ""
	}
	for _, addr := range addrs {
		if addr.String() != "" {
			return addr.String()
		}
	}
	return ""
}

type endpoint struct {
	id          string
	hostNic     *HostNic
	srcName     string
	portMapping []types.PortBinding // Operation port bindings
	dbIndex     uint64
	dbExists    bool
	sandboxKey  string
}

//HostNicDriver implements github.com/docker/go-plugins-helpers/network.Driver
type HostNicDriver struct {
	network     string
	excludeNics []string
	allnic      NicTable
	endpoints   map[string]*endpoint
	lock        sync.RWMutex
	nlh         *netlink.Handle
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
	d.lock.Lock()
	defer d.lock.Unlock()

	log.Debug("CreateEndpoint Called: [ %+v ]", r)
	log.Debug("r.Interface: [ %+v ]", r.Interface)

	hardwareAdddr := r.Interface.MacAddress
	hostNic := d.GetHostNicByHardwareAddr(hardwareAdddr)
	if hostNic == nil {
		hostNic = d.GetUnusedNic()
	}
	if hostNic == nil {
		return nil, errors.New("Can not find a free interface")
	}
	hostIfName := hostNic.NetInterface.Name

	// Generate a name for what will be the sandbox side pipe interface
	containerIfName, err := netutils.GenerateIfaceName(d.nlh, vethPrefix, vethLen)
	if err != nil {
		return err
	}

	// Generate and add the interface pipe host <-> sandbox
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: hostIfName, TxQLen: 0},
		PeerName:  containerIfName}
	if err = d.nlh.LinkAdd(veth); err != nil {
		return types.InternalErrorf("failed to add the host (%s) <=> sandbox (%s) pair interfaces: %v", hostIfName, containerIfName, err)
	}

	// Get the host side pipe interface handler
	host, err := d.nlh.LinkByName(hostIfName)
	if err != nil {
		return types.InternalErrorf("failed to find host side interface %s: %v", hostIfName, err)
	}
	defer func() {
		if err != nil {
			d.nlh.LinkDel(host)
		}
	}()

	// Get the sandbox side pipe interface handler
	sbox, err := d.nlh.LinkByName(containerIfName)
	if err != nil {
		return types.InternalErrorf("failed to find sandbox side interface %s: %v", containerIfName, err)
	}
	defer func() {
		if err != nil {
			d.nlh.LinkDel(sbox)
		}
	}()

	resp := &network.CreateEndpointResponse{
		Interface: &network.EndpointInterface{
			MacAddress: hostNic.HardwareAddr(),
			Address:    hostNic.Addr(),
		},
	}
	endpint := &endpoint{}

	// Store the sandbox side pipe interface parameters
	endpoint.srcName = containerIfName
	endpint.hostNic = hostNic
	endpint.id = r.EndpointID

	d.endpoints[endpint.id] = endpint
	hostNic.endpoint = endpint

	// Set the sbox's MAC if not provided. If specified, use the one configured by user, otherwise generate one based on IP.
	//if endpoint.macAddress == nil {
	//	endpoint.macAddress = electMacAddress(epConfig, endpoint.addr.IP)
	//	if err = ifInfo.SetMacAddress(endpoint.macAddress); err != nil {
	//		return err
	//	}
	//}

	// Up the host interface after finishing all netlink configuration
	//if err = d.nlh.LinkSetUp(host); err != nil {
	//	return fmt.Errorf("could not set link up for host interface %s: %v", hostIfName, err)
	//}

	//if endpoint.addrv6 == nil && config.EnableIPv6 {
	//	var ip6 net.IP
	//	network := n.bridge.bridgeIPv6
	//	if config.AddressIPv6 != nil {
	//		network = config.AddressIPv6
	//	}
	//
	//	ones, _ := network.Mask.Size()
	//	if ones > 80 {
	//		err = types.ForbiddenErrorf("Cannot self generate an IPv6 address on network %v: At least 48 host bits are needed.", network)
	//		return err
	//	}
	//
	//	ip6 = make(net.IP, len(network.IP))
	//	copy(ip6, network.IP)
	//	for i, h := range endpoint.macAddress {
	//		ip6[i+10] = h
	//	}
	//
	//	endpoint.addrv6 = &net.IPNet{IP: ip6, Mask: network.Mask}
	//	if err = ifInfo.SetIPAddress(endpoint.addrv6); err != nil {
	//		return err
	//	}
	//}
	//
	//if err = d.storeUpdate(endpoint); err != nil {
	//	return fmt.Errorf("failed to save bridge endpoint %s to store: %v", endpoint.id[0:7], err)
	//}

	return resp, nil
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
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debug("Join Called: [ %+v ]", r)
	endpoint := d.endpoints[r.EndpointID]

	if endpoint == nil {
		return nil, fmt.Errorf("Cannot find endpoint by id: %s", r.EndpointID)
	}

	endpoint.sandboxKey = r.SandboxKey
	d.endpoints[r.EndpointID] = r.SandboxKey
	resp := network.JoinResponse{
		InterfaceName:         network.InterfaceName{SrcName: endpoint.srcName, DstPrefix: containerVethPrefix},
		DisableGatewayService: true,
	}
	return &resp, nil
}
func (d *HostNicDriver) Leave(r *network.LeaveRequest) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debug("Leave Called: [ %+v ]", r)
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

func (d *HostNicDriver) findNics() {
	nics, err := net.Interfaces()

	if err == nil {
		for _, nic := range nics {
			if d.isExcludeNic(nic.Name) {
				continue
			}
			if _, ok := d.allnic[nic.HardwareAddr.String()]; ok {
				continue
			}
			hostNic := &HostNic{NetInterface: nic}
			log.Info("Find new nic: %+v ", nic)
			d.allnic[hostNic.HardwareAddr()] = hostNic
		}
	} else {
		log.Error("Get Interfaces error:%s", err.Error())
	}
}

func (d *HostNicDriver) isExcludeNic(nicName string) bool {
	if nicName == "lo" {
		return true
	}
	for _, nic := range d.excludeNics {
		if nicName == nic {
			return true
		}
	}
	return false
}

func (d *HostNicDriver) GetHostNicByHardwareAddr(hardwareAddr string) *HostNic {
	nic := d.getHostNicByHardwareAddr(hardwareAddr)
	if nic == nil {
		d.findNics()
		nic = d.getHostNicByHardwareAddr(hardwareAddr)
	}
	return nic
}

func (d *HostNicDriver) getHostNicByHardwareAddr(hardwareAddr string) *HostNic {
	if hostNic, ok := d.allnic[hardwareAddr]; ok {
		return hostNic
	}
	return nil
}

func (d *HostNicDriver) GetUnusedNic() *HostNic {
	nic := d.getUnusedNic()
	if nic == nil {
		d.findNics()
		nic = d.getUnusedNic()
	}
	return nic
}

func (d *HostNicDriver) getUnusedNic() *HostNic {
	for _, nic := range d.allnic {
		if nic.endpoint == nil {
			return nic
		}
	}
}
