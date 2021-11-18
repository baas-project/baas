# BAAS Project

BAAS (Baremetal as a Service) is a project aiming to facilitate research into operating systems by simplifying the process of downloading and syncing images to work servers. It allows for the registration of machines, the scheduling of flashing images and the uploading of images. The final goal of the project is to create a server where a research can mxs and match various personal or system images to easily test configurations. By allowing these to be run on servers they don't need to risk dataloss on their own machines and by automating it there also will not be long waiting times.

This project was made student developers from the Bsc Computer Science & Engineering at TU Delft and has been made with those machines in mind. Modifications to the project to have it run on other machines are more than welcome, but no garantuees can be made of it actually working on your target hardware..

## Project structure
At the moment the project three subprojects which are all separate self-contained programs which communicate to each other over well defined REST interfaces. The highest level abstraction, and the one most users will use, is the website interface which allows for the graphical management of the images. In turn, the website communicates over the network to a Control Server which deals with authentication, the storage of the images, and ensuring that the right configuration is sent to the server. On the server there is a management OS which takes configurations from the server, flashes them to the server and ensures that the server is always returned into a valid neutral state. A diagram of a typical interaction can be found in the [control server overview](control_server/index.md).

## Documentation
Structure of this manual is as follows: first a general overview is given of the general concepts inside of the BAAS project which is followed by in-depth explanation of each submodule. First, the control server is explored since it is the user facing software and will be of most interest to the vast majority of users. Afterwards, workings of the management OS are explained which contain most of the complexity and is the program that is actually run on the hardware.

## Installation instructions
Installing the BAAS project for development can be a bit tricky since, by definition, it requires running more than one machine. Luckily this can be circumvented by using a virtual machine, but using iPXE to serve to a virtual machine is not a typical use case. Adding to this complexity is the fact that every machine and operating system can react slightly differently, therefore if you have any problems with the installation please do not hesistate to file a Github issue.

### Required software
- Go
- cpio
- goimports (development)
- libvirt
- virt-manager or virt-viewer
- curl or POSTman (interacting with the control server)

### Virtual machine installation
Normally, you would use a virtual machine to create a client machine and run the server on your hardware. It is recommended to use libvirt for this, which is also what was used for initial development. You can set it up as follows:

1. First start virt-manager
2. Go to Edit->Connection Details and create a new network
3. Select NAT
4. Set the name as BAASNetwork and set "Forward To" as the network card you typically use (in my case wlan0).
5. Press on finish.

!!! warning "Firewalls can interfere with VM networks."
    If you have problems with connecting to the server, double check if there is not a firewall running in the background.
	Run the following command, after running it, you should be able to boot the virtual machine with the proper network settings.

```sh
virt-install --pxe --prompt --memory 2048 --name baas --disk size=30
--boot uefi,network,hd --network network=BAASNetwork --os-variant generic
```

Finally generate the management operating system image which is run on the client machine.

```sh
make management_initramfs
```

### Starting the control server
In `virt-manager` go to view and select Details, press on the light bulb and find the menu item called NIC. From there copy the MAC address and change the value in `control_server/main.go` to this MAC address. You can then run `make control_server` to run the control server.

### Scheduling the first boot
At boot the server will add the machine and hence the only thing left to do is ensuring that the system actually has images that it can boot. First create a user on the system, followed by the creation of an initial image and the downloading this image to disk. It is assumed that you have the `curl` and `jq` utilities installed and are running on a UNIX system.

```sh
	curl -X POST "localhost:4848/user" -H 'Content-Type: application/json' -d '{"name": "USER", "email": "EMAIL", "role": "user"}'
	UUID=$(curl -X POST "localhost:4848/user/USER/image" -H 'Content-Type: application/json' -d '{"name": "Test image", "DiskUUID": "/dev/sda"}' | jq .UUID | sed 's/\"//g')
	curl "localhost:4848/image/${UUID}/latest" --output /tmp/image.img
```

Running these commands creates an initial user called USER and a testing image which is downloaded on disk. This image can be modified in any arbitrary way or can be replaced with another file entirely. After creating the image the modified can be uploaded and scheduled for booting by running the following commands:

```sh
	VERSION=$(curl -X POST "localhost:4848/image/${UUID}" -H "Content-Type: multipart/form-data" -F "file=@/tmp/image.img" | awk '{print $4}')
	curl -X POST "localhost:4848/machine/[your MAC address]/boot" -H 'Content-Type: application/json' -d "{\"Version\": ${VERSION}, \"ImageUUID\": \"${UUID}\", \"update\": false}"
```

### Running the virtual machine
If you now boot your virtual machine it should now boot into the management OS, download your image, flash it into the disk and reboot into the image.
