module baas

go 1.15

require (
	github.com/baas-project/bmc v0.0.0-20200904230046-a5643220ab2a
	github.com/gebn/bmc v0.0.0-20200904230046-a5643220ab2a // indirect
	github.com/gorilla/mux v1.7.4
	github.com/krolaw/dhcp4 v0.0.0-20190909130307-a50d88189771
	github.com/sirupsen/logrus v1.4.2
	go.universe.tf/netboot v0.0.0-20200920222120-66e5fba6f663
)

replace github.com/baas-project/bmc => ../bmc
