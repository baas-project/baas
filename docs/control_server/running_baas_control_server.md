
# Baas control server usage guide

## prerequisites

1. Make sure the management kernel and initramfs have been generated and exist here: [`/control_server/static`](../../control_server/static)

### Running in docker

In the project root, run:

```bash
docker-compose up --build
```

or if the control server is already built:
```bash
docker-compose up
```

### Running outside of docker

In the project root, run:
```bash
sudo go run ./control_server
```
sudo is necesary here because BAAS listens on port 67 (dhcp)

## Usage

When the control server is running, any computer or virtual machine running on the same network can attempt to boot using PXE.
The control server will provide the machine with a management kernel based on its architecture. (Currently only tested on X86... #).

The first thing the management kernel will do when booted is to set up communication with the control server, after which it will initialize
the system based on directions given by the control server.


### Baas in a bridged virtual machine

When using BAAS to provision a local VM that's connected via a bridge network, it may be necessary to run this command:
```bash
/usr/bin/iptables -A FORWARD -p all -i virtual_machine_bridge -j ACCEPT
```
This is necessary because docker enables some iptable rules which apply to all bridge networks, including those of your
virtual machines. This stops network boot (and probably more things) from working inside the virtual machine
