# Rainier
Rainier is a container network interface (CNI) plugin that creates one network interface for each container and attach it to an OVS bridge

## How to use
### Prerequisite
Install `golang`, `govendor`, `kubernetes`, and `kubernetes-cni`

### Fetch rainier source code
```bash
go get -u github.com/charlesmchan/rainier
```

### Build
```bash
cd $GOPATH/src/github.com/charlesmchan/rainier
govendor sync
go build rainier.go
```
### Link binary and configuration to CNI folder
```bash
sudo mkdir -p /etc/cni/net.d
sudo mkdir -p /opt/cni/bin
sudo ln -s config /etc/cni/net.d/rainier.conf
sudo ln -s rainier /opt/cni/bin/rainier
```

### Test with sample service
```bash
sudo kubectl create -f service.yaml
```
`service.yaml` will spawn 3 `busybox` containers. All of them should get an IP address and should be able to ping one another if OVS is set to standalone mode

## Note
Kubernetes does not take DNS configuration returned from CNI. We need to configure DNS in the Kubernetes pod configuration

## Todo
- Test cases

## How it is named
I have a bad sense of naming a project and I was eating rainier cherries while coding

## Acknowledgement
Special thanks to @hwchiu

## Reference
- [CNI](https://github.com/containernetworking/cni)
- [CNI plugin](https://github.com/containernetworking/plugins)
- [Go OpenvSwitch library](https://github.com/digitalocean/go-openvswitch)
