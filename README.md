# Rainier CNI (Container Network Interface)
This is just a production of me practicing Go and CNI. It likely does *NOT* work.

## What it does
Create one network interface for each container and attach it to an OVS bridge

## How it is named
I have a bad sense of naming a project and I was eating rainier cherries while coding

## Note
Kubernetes does not take DNS configuration returned by CNI. Need to configure DNS in POD configuration

## Todo
- Test cases

## Reference
- [CNI](https://github.com/containernetworking/cni)
- [CNI plugin](https://github.com/containernetworking/plugins)
- [Go OpenvSwitch library](https://github.com/digitalocean/go-openvswitch)
