
# How to boot over pxe using pixiecore

## prerequisites

1. An existing dhcp server (!!!)

## Booting

1. run the control server with the pixiecore api
2. run pixiecore: `pixiecore api http://127.0.0.1:4242`
3. start your machine (same network as the dhcp server and control server)
4. pray

# dhcpd config:

option domain-name-servers 8.8.8.8, 8.8.4.4;
option subnet-mask 255.255.255.0;
option routers 172.19.0.1; <-- docker bridge network
subnet 172.19.0.0 netmask 255.255.255.0 { <-- docker bridge network with 0
      range 172.19.0.100 172.19.0.250; <-- some range
}
