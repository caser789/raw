// Package raw enables reading and writing data at the device driver level for
// a network interface.
package raw

import (
    "errors"
    "net"
)

var (
    // ErrNotImplemented is returned when certain functionality is not yet
    // implemented for the host operating system.
    ErrNotImplemented = errors.New("not implemented")
)

// Addr is a network address which can be used to contact other machines, using
// their hardware addresses.
type Addr struct {
    HardwareAddr net.HardwareAddr
}

// Network returns the address's network name, "raw".
func (a Addr) Network() string {
    return "raw"
}

// String returns the address's hardware address.
func (a Addr) String() string {
    return a.HardwareAddr.String()
}

// ListenPacket creates a net.PacketConn which can be used to send and receive
// data at the network interface device driver level.
//
// ifi specifies the network interface which will be used to send and receive
// data. socket specifies the socket type to be used, such as syscall.SOCK_RAW
// or syscall.SOCK_DGRAM. proto specifies the protocol which should be captured
// and transimitted. proto is automattically converted to network byte
// order (big endian), akin to the htons() function in C.
func ListenPacket(ifi *net.Interface, proto int) (net.PacketConn, error) {
    return listenPacket(ifi, proto)
}

// htons converts a short (uint16) from host-to-network byte order.
func htons(i uint16) uint16 {
    return (i<<8)&0xff00 | i>>8
}

type timeoutError struct{}

func (e *timeoutError) Error() string { return "i/o timeout" }
func (e *timeoutError) Timeout() bool { return true }
func (e *timeoutError) Temporary() bool { return true }
