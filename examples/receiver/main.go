package main

import "net"
import "log"
import "github.com/caser789/raw"
import "github.com/mdlayher/ethernet"

const etherType = 0xcccc

// const ifName = "enp2s0"
const ifName = "wlp3s0"

func main() {
	log.Println("start receiver")

	ifi, err := net.InterfaceByName(ifName)
	if err != nil {
		log.Fatalf("failed to open interface: %v", err)
	}
	log.Println(ifi)

	c, err := raw.ListenPacket(ifi, etherType)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println(c)
	defer c.Close()

	b := make([]byte, ifi.MTU)
	var f ethernet.Frame

	for {
		n, addr, err := c.ReadFrom(b)
		if err != nil {
			log.Fatalf("failed to receive message: %v", err)
		}

		if err := (&f).UnmarshalBinary(b[:n]); err != nil {
			log.Fatalf("failed to unmarshal ethernet frame: %v", err)
		}

		log.Printf("[%s] %s", addr.String(), string(f.Payload))
	}
}
