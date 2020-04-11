// +build !linux

package raw

import (
    "net"
    "time"
)

const (
    ProtocolARP Protocol = 0
)

var (
    // Must implement net.PacketConn at compile-time.
    _ net.PacketConn = &packetConn{}
)

// packetConn is the generic implementation of net.PacketConn for this
type packetConn struct { }

// listenPacket creates a net.PacketConn which can be used to send and receive
// data at the device driver level.
//
// ifi specifies the network interface which will be used to send and receive
// data. socket specifies the socket type to be used, such as syscall.SOCK_RAW
// or syscall.SOCK_DGRAM. proto specifies the protocol which should be
// captured and transmitted. proto is automatically converted to network byte
// order (big endian), akin to the htons() function in C.
func listenPacket(ifi *net.Interface, proto Protocol) (*packetConn, error) {
    return nil, ErrNotImplemented
}

// ReadFrom implements the net.PacketConn.ReadFrom method.
func (p *packetConn) ReadFrom(b []byte) (int, net.Addr, error) {
    return 0, nil, ErrNotImplemented
}

// WriteTo implements the net.PacketConn.WriteTo method.
func (p *packetConn) WriteTo(b []byte, addr net.Addr) (int, error) {
    return 0, ErrNotImplemented
}

// CLose closes the connection
func (p *packetConn) Close() error {
    return ErrNotImplemented
}

// LocalAddr returns the local network address.
func (p *packetConn) LocalAddr() net.Addr {
    return nil
}

// SetDeadline implements the net.PacketConn.SetDeadline method
func (p *packetConn) SetDeadline(t time.Time) error {
    return ErrNotImplemented
}

// SetReadDeadline implements the net.PacketConn.SetReadDeadline method
func (p *packetConn) SetReadDeadline(t time.Time) error {
    return ErrNotImplemented
}

// SetWriteDeadline implements the net.PacketConn.SetWriteDeadline method
func (p *packetConn) SetWriteDeadline(t time.Time) error {
    return ErrNotImplemented
}
