
# BAAS

Baremetal As A Service, or abbreviated to BAAS is a project done for the TU Delft to facilitate operating systems
research on high-powered machines, by scheduling the access researches have to these machines. This system makes sure
each researcher can have full access to the machines in their timeslot, being able to load their own (custom) kernel and have
access to all hardware. BAAS makes sure the system is reset a well-known state after every job, to make sure these custom kernels
have not messed up the system enough to make further research on them impossible.

Disclaimer: this system is custom-built for the needs of the TU Delft. It is open source, and written to be extensible,
but it is unlikely that it will completely fit your needs without changes. We are open to pull requests,
but we might not implement suggestions ourselves which are outside the scope of the requirements for the TU Delft.

# Documentation

Some documentation about the systems, and building these systems is provided [here](https://baas-project.github.io/baas/)

# Install

## Required software
- Go
- cpio
- goimports
- libvirt
- virt-manager/virt-viewer

## Installation of the virtual machine
This software works on a client a server-model, where there is a
central control server which offers the management OS to one or multiple 
clients. These are two entirely different systems and hence both
testing as well as developing must be done on two separate machines.

### Client machine
Normally, you would use a virtual machine to create a client machine
and run the server on your hardware. It is recommended to use libvirt
for this, which is also what was used for initial development. You can
set it up as follows:

First start virt-manager, go to Edit->Connection Details and create a
new network, select NAT, set the name as BAASNetwork and set "Forward
To" as the network card you typically use (in my case wlan0). Press on
finish.

> :warning: If you have problems with connecting to the server, 
>  double check if there is not a firewall running in the background. 
Run the following command, after running it, you should be able to
boot the virtual machine with the proper network settings.

```sh
virt-install --pxe --prompt --memory 2048 --name baas --disk size=30
--boot uefi,network,hd --network network=BAASNetwork --os-variant generic
```

Finally generate the management operating system which is run on the 
client machine. 

```sh
make management_initramfs
```

### Control server
In `virt-manager` go to view and select Details, press on the light
bulb and find the menu item called NIC. From there copy the MAC
address and change the value in control_server/main.go to this IP
address. You can then run `make control_server` to run the control
server and reboot the virtual machine. If all is well, it should now
boot into the management operating system.

# License

// TODO
