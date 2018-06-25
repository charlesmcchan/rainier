package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/digitalocean/go-openvswitch/ovs"
)

const DefaultMTU = 1500
const HostInterfacesFilePath = "/tmp/rainier.json"

var hostInterfaces = make(map[string]interface{})

type RainierConfig struct {
	types.NetConf
	PublicBridgeName string `json:"publicBridgeName"`
	// TODO Add private interface
	// PrivateBridgeName string `json:"privateBridgeName"`
}

func cmdAdd(args *skel.CmdArgs) error {
	config := &RainierConfig{}
	if err := json.Unmarshal(args.StdinData, config); err != nil {
		return err
	}

	// Create OVS bridges
	if err := createOvsBr(config.PublicBridgeName); err != nil {
		return err
	}
	// TODO Add private interface
	// if err := createOvsBr(config.PrivateBridgeName); err != nil {
	// 	return err
	// }

	// Get name space
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// Create veth
	hostInterface, _, err := createVeth(netns, args.IfName)
	if err != nil {
		return err
	}

	// Add port to OVS
	if err := addOvsPort(config.PublicBridgeName, hostInterface.Name); err != nil {
		return err
	}

	// Update JSON file
	readHostInterfacesFromFile()
	hostInterfaces[args.ContainerID] = hostInterface.Name
	writeHostInterfacesToFile()

	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	config := &RainierConfig{}
	if err := json.Unmarshal(args.StdinData, config); err != nil {
		return err
	}

	// Update JSON file and remove port from OVS
	readHostInterfacesFromFile()
	hostIfName := hostInterfaces[args.ContainerID].(string)
	if err := deleteOvsPort(config.PublicBridgeName, hostIfName); err != nil {
		return err
	}
	delete(hostInterfaces, args.ContainerID)
	writeHostInterfacesToFile()

	return nil
}

func cmdGet(args *skel.CmdArgs) error {
	return fmt.Errorf("cmdGet is not implemented")
}

func createVeth(netns ns.NetNS, ifName string) (*current.Interface, *current.Interface, error) {
	contIface := &current.Interface{}
	hostIface := &current.Interface{}

	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, DefaultMTU, hostNS)
		if err != nil {
			return err
		}
		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = netns.Path()
		hostIface.Name = hostVeth.Name
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return hostIface, contIface, nil
}

func createOvsBr(bridgeName string) error {
	protocols := []string{ovs.ProtocolOpenFlow13}
	client := ovs.New(
		ovs.Sudo(),
		ovs.Protocols(protocols),
	)

	if err := client.VSwitch.AddBridge(bridgeName); err != nil {
		return fmt.Errorf("Failed to add bridge %s. Error = %s", bridgeName, err)
	}
	return nil
}

func addOvsPort(bridgeName string, hostIfName string) error {
	protocols := []string{ovs.ProtocolOpenFlow13}
	client := ovs.New(
		ovs.Sudo(),
		ovs.Protocols(protocols),
	)

	if err := client.VSwitch.AddPort(bridgeName, hostIfName); err != nil {
		return fmt.Errorf("Failed to add port %s to bridge %s. Error = %s", hostIfName, bridgeName, err)
	}
	return nil
}

func deleteOvsPort(bridgeName string, hostIfName string) error {
	protocols := []string{ovs.ProtocolOpenFlow13}
	client := ovs.New(
		ovs.Sudo(),
		ovs.Protocols(protocols),
	)

	if err := client.VSwitch.DeletePort(bridgeName, hostIfName); err != nil {
		fmt.Println("%s %s", bridgeName, hostIfName)
		return fmt.Errorf("Failed to delete port %s from bridge %s. Error = %s", hostIfName, bridgeName, err)
	}
	return nil
}

func readHostInterfacesFromFile() error {
	jsonByte, err := ioutil.ReadFile(HostInterfacesFilePath)
	if err == nil {
		if err := json.Unmarshal(jsonByte, &hostInterfaces); err != nil {
			return fmt.Errorf("Fail to decode host interface JSON")
		}
	}
	return nil
}

func writeHostInterfacesToFile() error {
	jsonByte, err := json.Marshal(hostInterfaces)
	if err != nil {
		return fmt.Errorf("Fail to encode host interface JSON")
	}
	if err := ioutil.WriteFile(HostInterfacesFilePath, jsonByte, 0644); err != nil {
		return fmt.Errorf("Fail to write host interface JSON")
	}
	return nil
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	about := "Rainier CNI"
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, about)
}
