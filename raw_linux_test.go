// +build linux

package raw

import (
    "testing"
    "errors"
    "net"
    "sync/atomic"
    "syscall"
    "time"
    "bytes"
)


// Test for errors which occur when attempting to set socket to
// nonblocking mode.

type errSetNonblockSocket struct {
    err error
    noopSocket
}

func (s *errSetNonblockSocket) SetNonblock(nonblocking bool) error { return s.err }

func Test_newPacketConnSetNonblockError(t *testing.T) {
    fooErr := errors.New("foo")

    _, err := newPacketConn(
        &net.Interface{},
        &errSetNonblockSocket{
            err: fooErr,
        },
        0,
        &testSleeper{},
    )
    if want, got := fooErr, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }
}

// Test to ensure that nonblocking mode is always true fro sockets

type setNonblockSocket struct {
    nonblocking bool
    noopSocket
}

func (s *setNonblockSocket) SetNonblock(nonblocking bool) error {
    s.nonblocking = nonblocking
    return nil
}

func Test_newPacketConnSetNonblock(t *testing.T) {
    s := &setNonblockSocket{}

    _, err := newPacketConn(
        &net.Interface{},
        s,
        0,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    if want, got := true, s.nonblocking; want != got {
        t.Fatalf("unexpected nonblocking boolean:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for errors which occur while attempting to bind socket.

type errBindSocket struct {
    err error
    noopSocket
}

func (s *errBindSocket) Bind(sa syscall.Sockaddr) error { return s.err }

func Test_newPacketConnBindError(t *testing.T) {
    fooErr := errors.New("foo")

    _, err := newPacketConn(
        &net.Interface{},
        &errBindSocket{
            err: fooErr,
        },
        0,
        &testSleeper{},
    )
    if want, got := fooErr, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }
}

// Test to ensure that socket is bound with correct sockaddr_ll information

type bindSocket struct{
    bind syscall.Sockaddr
    noopSocket
}

func (s *bindSocket) Bind(sa syscall.Sockaddr) error {
    s.bind = sa
    return nil
}

func Test_newPacketConnBind(t *testing.T) {
    s := &bindSocket{}

    ifIndex := 1
    protocol := uint16(1)

    _, err := newPacketConn(
        &net.Interface{
            Index: ifIndex,
        },
        s,
        protocol,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    sall, ok := s.bind.(*syscall.SockaddrLinklayer)
    if !ok {
        t.Fatalf("bind sockaddr has incorrect type: %T", s.bind)
    }

    if want, got := ifIndex, sall.Ifindex; want != got {
        t.Fatalf("unexpected network interface index:\n- want: %v\n- got: %v", want, got)
    }
    if want, got := protocol, sall.Protocol; want != got {
        t.Fatalf("unexpected protocol:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for errors which occur immediately when calling recvfrom on a socket

type errRecvfromSocket struct {
    err error
    noopSocket
}

func (s *errRecvfromSocket) Recvfrom(p []byte, flags int) (int, syscall.Sockaddr, error) {
    return 0, nil, s.err
}

func Test_packetConnReadFromRecvfromError(t *testing.T) {
    fooErr := errors.New("foo")

    p, err := newPacketConn(
        &net.Interface{},
        &errRecvfromSocket{
            err: fooErr,
        },
        0,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    _, _, err = p.ReadFrom(nil)
    if want, got := fooErr, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for errors which occur after several retries while attempting to
// recvfrom on a socket

type errRetryNRecvfromSocket struct{
    n int
    try int
    err error
    noopSocket
}

func (s *errRetryNRecvfromSocket) Recvfrom(p []byte, flags int) (int, syscall.Sockaddr, error) {
    if s.try == s.n {
        return 0, nil, s.err
    }

    s.try++
    return 0, nil, syscall.EAGAIN
}

func Test_packetConnReadFromRecvfromRetryNError(t *testing.T) {
    fooErr := errors.New("foo")

    ts := &testSleeper{}

    const n = 5

    p, err := newPacketConn(
        &net.Interface{},
        &errRetryNRecvfromSocket{
            n: n,
            err: fooErr,
        },
        0,
        ts,
    )
    if err != nil {
        t.Fatal(err)
    }

    _, _, err = p.ReadFrom(nil)
    if want, got := fooErr, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }

    if want, got := n*(2*time.Millisecond), time.Duration(ts.slept); want != got {
        t.Fatalf("unexpected mock sleep time:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for incorrect swockaddr type after recvfrom on a socket

type addrRecvfromSocket struct {
    addr syscall.Sockaddr
    noopSocket
}

func (s *addrRecvfromSocket) Recvfrom(p []byte, flags int) (int, syscall.Sockaddr, error) {
    return 0, s.addr, nil
}

func Test_packetConnReadFromRecvfromInvalidSockaddr(t *testing.T) {
    p, err := newPacketConn(
        &net.Interface{},
        &addrRecvfromSocket{
            addr: &syscall.SockaddrInet4{},
        },
        0,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    _, _, err = p.ReadFrom(nil)
    if want, got := syscall.EINVAL, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for malformed hardware address after recvfrom on a socket

func Test_packetConnReadFromRecvfromInvalidHardwareAddr(t *testing.T) {
    p, err := newPacketConn(
        &net.Interface{},
        &addrRecvfromSocket{
            addr: &syscall.SockaddrLinklayer{
                Halen: 5,
            },
        },
        0,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    _, _, err = p.ReadFrom(nil)
    if want, got := syscall.EINVAL, err; want != got {
        t.Fatalf("unexpected error:\n- want: %v\n- got: %v", want, got)
    }
}

// Test for a correct ReadFrom with data and address

type recvfromSocket struct {
    p []byte
    addr syscall.Sockaddr
    noopSocket
}

func (s *recvfromSocket) Recvfrom(p []byte, flags int) (int, syscall.Sockaddr, error) {
    copy(p, s.p)
    return len(s.p), s.addr, nil
}

func Test_packetConnReadFromRecvfromOK(t *testing.T) {
    const wantN = 4
    data := []byte{0, 1, 2, 3}
    deadbeefHW := net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad}

    p, err := newPacketConn(
        &net.Interface{},
        &recvfromSocket{
            p: data,
            addr: &syscall.SockaddrLinklayer{
                Halen: 6,
                Addr: [8]byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0x00, 0x00},
            },
        },
        0,
        &testSleeper{},
    )
    if err != nil {
        t.Fatal(err)
    }

    buf := make([]byte, 8)
    n, addr, err := p.ReadFrom(buf)
    if err != nil {
        t.Fatal(err)
    }

    raddr, ok := addr.(*Addr)
    if !ok {
        t.Fatalf("read sockaddr has incorrect type: %T", addr)
    }
    if want, got := deadbeefHW, raddr.HardwareAddr; !bytes.Equal(want, got) {
        t.Fatalf("unexpected hardware address:\n- want: %v\n- got: %v", want, got)
    }

    if want, got := wantN, n; want != got {
        t.Fatalf("unexpected data length:\n- want: %v\n- got: %v", want, got)
    }

    if want, got := data, buf[:n]; !bytes.Equal(want, got) {
        t.Fatalf("unexpected data:\n- want: %v\n- got: %v", want, got)
    }
}

// testSleeper is a sleeper implementation which atomically increments a
// couter to indicate how long it has slept
type testSleeper struct {
    slept int64
}

func (t *testSleeper) Sleep(d time.Duration) {
    atomic.AddInt64(&t.slept, int64(d))
}

// noopSocket is a socket implementation which noops every operation. It is
// the basis for more specific socket implementations
type noopSocket struct{}

func (_ noopSocket) Bind(sa syscall.Sockaddr) error { return nil }
func (_ noopSocket) Close() error { return nil }
func (_ noopSocket) Recvfrom(p []byte, flags int) (int, syscall.Sockaddr, error) { return 0, nil, nil }
func (_ noopSocket) Sendto(p []byte, flags int, to syscall.Sockaddr) error { return nil }
func (_ noopSocket) SetNonblock(nonblocking bool) error { return nil }
