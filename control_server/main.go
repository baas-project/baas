package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/krolaw/dhcp4"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)


var (
	port = flag.Int("port", 4242, "Port to listen on")
	static = flag.String("static", "./static/", "Static file dir to server under /static/")
)

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/v1/boot/{mac}", api)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(*static))))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":"+strconv.Itoa(*port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go listenDhcp()

	log.Fatal(srv.ListenAndServe())
}

type DhcpHandler struct {

}

func (DhcpHandler) ServeDHCP(req dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	arch := options[dhcp4.OptionClientArchitecture]
	mac := req.CHAddr()
	num := req.XId()

	fmt.Printf("Architecture: %v, mac address: %v (num: %v)", arch, mac, num)

	return dhcp4.Packet{}
}

func listenDhcp() {
	err := dhcp4.ListenAndServe(DhcpHandler{})
	if err != nil {
		println(err.Error())
	}

	//for {
	//	conn, err := net.ListenUDP("udp", &net.UDPAddr{
	//		IP:   net.ParseIP("0.0.0.0"),
	//		Port: 67,
	//	})
	//
	//	if err != nil {
	//		println(err.Error())
	//		continue
	//	}
	//
	//	println("Received DHCP packet")
	//
	//	bytes := make([]byte, 512)
	//	_, addr, err := conn.ReadFromUDP(bytes)
	//	fmt.Printf("address: %v", addr)
	//	if err != nil {
	//		println(err.Error())
	//		continue
	//	}
	//
	//	println("Read dhcp packet")
	//	println(string(bytes))
	//
	//	err = conn.Close()
	//	if err != nil {
	//		println(err.Error())
	//		continue
	//	}
	//
	//	dhcp4.pa
	//}
}

func api(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving boot config for %s", filepath.Base(r.URL.Path))
	resp := struct {
		K string	`json:"kernel"`
		I []string	`json:"initrd"`
		//C string	`json:"cmdline"`
	}{
		K: "http://localhost:4242/static/vmlinuz/",
		I: []string{
			"http://localhost:4242/static/initramfs",
		},
		//C: "squashfs,sd-mod,usb-storage quiet nomodeset",
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		panic(err)
	}
}
