package main

import "net"
import "log"
import "github.com/caser789/raw/src/raw"
import "github.com/mdlayher/ethernet"
// import "github.com/mdlayher/raw"

const etherType = 0xcccc

func main() {
    log.Println("start")
    ifi, err := net.InterfaceByName("enp2s0")
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

    addr := &raw.Addr{HardwareAddr: ethernet.Broadcast}
    log.Printf("ethernet.Broadcast %T", ethernet.Broadcast)
    log.Println(ethernet.Broadcast)

    source := ifi.HardwareAddr
    msg := "Hello world"

    f := &ethernet.Frame{
        Destination: ethernet.Broadcast,
        Source: source,
        EtherType: etherType,
        Payload: []byte(msg),
    }
    b, err := f.MarshalBinary()
    if err != nil {
        log.Fatalf("failed to marshal frame: %v", err)
    }

    if _, err := c.WriteTo(b, addr); err != nil {
        log.Fatalf("failed to write frame: %v", err)
    }
}
