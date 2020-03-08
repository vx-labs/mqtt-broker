package gossip

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65536

	// udpRecvBufSize is a large buffer size that we attempt to set UDP
	// sockets to in order to handle a large volume of messages.
	udpRecvBufSize = 2 * 1024 * 1024
)

// NetTransport is a Transport implementation that uses connectionless UDP for
// packet operations, and ad-hoc TCP connections for stream operations.
type NetTransport struct {
	packetCh     chan *memberlist.Packet
	streamCh     chan net.Conn
	logger       *zap.Logger
	wg           sync.WaitGroup
	tcpListeners []net.Listener
	udpListeners []net.PacketConn
	shutdown     int32
	service      Service
}

// newNetTransport returns a net transport with the given configuration. On
// success all the network listeners will be created and listening.
func newNetTransport(service Service, logger *zap.Logger) (*NetTransport, error) {

	// Build out the new transport.
	var ok bool
	t := NetTransport{
		packetCh: make(chan *memberlist.Packet),
		streamCh: make(chan net.Conn),
		logger:   logger,
		service:  service,
	}

	// Clean up listeners if there's an error.
	defer func() {
		if !ok {
			t.Shutdown()
		}
	}()

	// Build all the TCP and UDP listeners.
	port := service.BindPort()
	addr := "0.0.0.0"

	tcpLn, err := service.ListenTCP()
	if err != nil {
		return nil, fmt.Errorf("Failed to start TCP listener on %q port %d: %v", addr, port, err)
	}
	t.tcpListeners = append(t.tcpListeners, tcpLn)

	// If the config port given was zero, use the first TCP listener
	// to pick an available port and then apply that to everything
	// else.
	if port == 0 {
		port = tcpLn.Addr().(*net.TCPAddr).Port
	}

	udpLn, err := service.ListenUDP()
	if err != nil {
		return nil, fmt.Errorf("Failed to start UDP listener on %q port %d: %v", addr, port, err)
	}
	t.udpListeners = append(t.udpListeners, udpLn)

	// Fire them up now that we've been able to create them all.
	for i := 0; i < 1; i++ {
		t.wg.Add(2)
		go t.tcpListen(t.tcpListeners[i])
		go t.udpListen(t.udpListeners[i])
	}

	ok = true
	return &t, nil
}

func (t *NetTransport) GetAutoBindPort() int {
	return t.service.BindPort()
}

// See Transport.
func (t *NetTransport) FinalAdvertiseAddr(_ string, _ int) (net.IP, int, error) {
	advAddress := t.service.Address()
	address, portStr, err := net.SplitHostPort(advAddress)
	if err != nil {
		return nil, 0, err
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		return nil, 0, err
	}
	ip := net.ParseIP(address)
	if ip == nil {
		t.logger.Fatal("failed to parse IP address", zap.String("provided_address", advAddress))
	}
	return ip, int(port), nil
}

// See Transport.
func (t *NetTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return time.Time{}, err
	}

	// We made sure there's at least one UDP listener, so just use the
	// packet sending interface on the first one. Take the time after the
	// write call comes back, which will underestimate the time a little,
	// but help account for any delays before the write occurs.
	_, err = t.udpListeners[0].WriteTo(b, udpAddr)
	return time.Now(), err
}

// See Transport.
func (t *NetTransport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

// See Transport.
func (t *NetTransport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{Timeout: timeout}
	return dialer.Dial("tcp", addr)
}

// See Transport.
func (t *NetTransport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

// See Transport.
func (t *NetTransport) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&t.shutdown, 1)

	// Rip through all the connections and shut them down.
	for _, conn := range t.tcpListeners {
		conn.Close()
	}
	for _, conn := range t.udpListeners {
		conn.Close()
	}

	// Block until all the listener threads have died.
	t.wg.Wait()
	return nil
}

// tcpListen is a long running goroutine that accepts incoming TCP connections
// and hands them off to the stream channel.
func (t *NetTransport) tcpListen(tcpLn net.Listener) {
	defer t.wg.Done()

	// baseDelay is the initial delay after an AcceptTCP() error before attempting again
	const baseDelay = 5 * time.Millisecond

	// maxDelay is the maximum delay after an AcceptTCP() error before attempting again.
	// In the case that tcpListen() is error-looping, it will delay the shutdown check.
	// Therefore, changes to maxDelay may have an effect on the latency of shutdown.
	const maxDelay = 1 * time.Second

	var loopDelay time.Duration
	for {
		iconn, err := tcpLn.Accept()
		conn := iconn.(*net.TCPConn)
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			if loopDelay == 0 {
				loopDelay = baseDelay
			} else {
				loopDelay *= 2
			}

			if loopDelay > maxDelay {
				loopDelay = maxDelay
			}

			t.logger.Error("error accepting TCP connection", zap.Error(err))
			time.Sleep(loopDelay)
			continue
		}
		// No error, reset loop delay
		loopDelay = 0

		t.streamCh <- conn
	}
}

// udpListen is a long running goroutine that accepts incoming UDP packets and
// hands them off to the packet channel.
func (t *NetTransport) udpListen(udpLn net.PacketConn) {
	defer t.wg.Done()
	for {
		// Do a blocking read into a fresh buffer. Grab a time stamp as
		// close as possible to the I/O.
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := udpLn.ReadFrom(buf)
		ts := time.Now()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}
			t.logger.Error("error reading UDP packet", zap.Error(err))
			continue
		}

		// Check the length - it needs to have at least one byte to be a
		// proper message.
		if n < 1 {
			t.logger.Error("UDP packet too short", zap.Int("packet_size", len(buf)))
			continue
		}

		// Ingest the packet.
		metrics.IncrCounter([]string{"memberlist", "udp", "received"}, float32(n))
		t.packetCh <- &memberlist.Packet{
			Buf:       buf[:n],
			From:      addr,
			Timestamp: ts,
		}
	}
}
