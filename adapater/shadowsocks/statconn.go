package shadowsocks

import (
	"net"
	"sync/atomic"
)

type StatConn struct {
	net.Conn
	Stat *Stat
}

func WrapConn(c net.Conn, stat *Stat) *StatConn {
	conn := &StatConn{Conn: c, Stat: stat}
	return conn
}

func (c *StatConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	atomic.AddUint64(&c.Stat.TcpRecvBytes, uint64(n))

	return n, err
}

func (c *StatConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if err != nil {
		return n, err
	}
	atomic.AddUint64(&c.Stat.TcpSendBytes, uint64(n))

	return n, nil
}

func (c *StatConn) Close() error {
	return c.Conn.Close()
}

type StatPacketConn struct {
	net.PacketConn
	Stat *Stat
}

func WrapPacketConn(pc net.PacketConn, stat *Stat) *StatPacketConn {
	conn := &StatPacketConn{PacketConn: pc, Stat: stat}
	return conn
}

func (pc *StatPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := pc.PacketConn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}
	atomic.AddUint64(&pc.Stat.UdpRecvBytes, uint64(n))
	atomic.AddUint64(&pc.Stat.UdpRecvPackets, 1)

	return n, addr, err
}

func (pc *StatPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := pc.PacketConn.WriteTo(b, addr)
	if err != nil {
		return n, err
	}
	atomic.AddUint64(&pc.Stat.UdpSendBytes, uint64(n))
	atomic.AddUint64(&pc.Stat.UdpSendPackets, 1)

	return n, nil
}

func (pc *StatPacketConn) Close() error {
	return pc.PacketConn.Close()
}
