// Package raw enables reading and writing data at the device driver level for
// a network interface.
package raw

import (
	"errors"
	"net"
	"time"
)

var (
	// ErrNotImplemented is returned when certain functionality is not yet
	// implemented for the host operating system.
	ErrNotImplemented = errors.New("not implemented")
)

var _ net.Addr = &Addr{}

// Addr is a network address which can be used to contact other machines, using
// their hardware addresses.
type Addr struct {
	HardwareAddr net.HardwareAddr
}

// Network returns the address's network name, "raw".
func (a *Addr) Network() string {
	return "raw"
}

// String returns the address's hardware address.
func (a *Addr) String() string {
	return a.HardwareAddr.String()
}

var _ net.PacketConn = &Conn{}

// Conn is an implementation of the net.PacketConn interface which ca send
// and receive data at the network interface device driver lvel
type Conn struct {
	// packetConn is the operating system-specific implementation of
	// a raw connection
	p *packetConn
}

// ReadFrom implements the net.PacketConn ReadFrom method
func (c *Conn) ReadFrom(b []byte) (int, net.Addr, error) {
	return c.p.ReadFrom(b)
}

// WriteTo
func (c *Conn) WriteTo(b []byte, addr net.Addr) (int, error) {
	return c.p.WriteTo(b, addr)
}

func (c *Conn) Close() error {
	return c.p.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.p.LocalAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.p.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.p.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.p.SetWriteDeadline(t)
}

// A protocol is a network protocol constant which id the type of
// traffic a raw socket should send and receive
type Protocol uint16

// ListenPacket creates a net.PacketConn which can be used to send and receive
// data at the network interface device driver level.
//
// ifi specifies the network interface which will be used to send and receive
// data. socket specifies the socket type to be used, such as syscall.SOCK_RAW
// or syscall.SOCK_DGRAM. proto specifies the protocol which should be captured
// and transimitted. proto is automattically converted to network byte
// order (big endian), akin to the htons() function in C.
func ListenPacket(ifi *net.Interface, proto Protocol) (*Conn, error) {
	p, err := listenPacket(ifi, proto)
	if err != nil {
		return nil, err
	}

	return &Conn{
		p: p,
	}, nil
}

// htons converts a short (uint16) from host-to-network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
