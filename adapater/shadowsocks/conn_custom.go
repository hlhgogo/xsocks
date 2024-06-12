package shadowsocks

import (
	"github.com/juju/ratelimit"
	"net"
	"strconv"
	"time"
)

const (
	OpRateLimit = 1 << iota
	OpReportTCP
)

type ConnCustom struct {
	net.Conn
	OpFlag   int
	TxBucket *ratelimit.Bucket
	RxBucket *ratelimit.Bucket
	tcpInfo  tcpInfo
}

type tcpInfo struct {
	enable bool
	start  time.Time
	rx     int64
	tx     int64
	rp     func(client_ip, tx, rx string, d time.Duration)
}

type TcpReporter func(client_ip, tx, rx string, d time.Duration)

func UpgradeTCPConn(c *net.TCPConn) *ConnCustom {
	return &ConnCustom{Conn: c, OpFlag: 0}
}

func (c *ConnCustom) EnableTCPReporter(rp TcpReporter) {
	c.OpFlag = c.OpFlag | OpReportTCP
	c.tcpInfo = tcpInfo{
		enable: true,
		start:  time.Now(),
		rx:     0,
		tx:     0,
		rp:     rp,
	}
}

func (c *ConnCustom) isFlagOn(Flag int) bool {
	if c.OpFlag&Flag != 0 {
		return true
	}
	return false
}

func (c *ConnCustom) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	if c.isFlagOn(OpRateLimit) && c.RxBucket != nil {
		c.RxBucket.Wait(int64(n * 8))
	}
	if c.isFlagOn(OpReportTCP) && c.tcpInfo.enable {
		c.tcpInfo.tx += int64(n)
	}
	return n, err
}

func (c *ConnCustom) Write(b []byte) (int, error) {
	if c.isFlagOn(OpRateLimit) && c.TxBucket != nil {
		c.TxBucket.Wait(int64(len(b) * 8))
	}
	if c.isFlagOn(OpReportTCP) && c.tcpInfo.enable {
		c.tcpInfo.rx += int64(len(b))
	}
	return c.Conn.Write(b)
}

func (c *ConnCustom) Close() error {
	if c.isFlagOn(OpReportTCP) && c.tcpInfo.enable && c.tcpInfo.rp != nil &&
		c.tcpInfo.rx > 0 && c.tcpInfo.tx > 0 {
		c.tcpInfo.rp(c.RemoteAddr().(*net.TCPAddr).IP.String(),
			strconv.FormatInt(c.tcpInfo.tx, 10),
			strconv.FormatInt(c.tcpInfo.rx, 10),
			time.Since(c.tcpInfo.start))
	}
	return c.Conn.Close()
}

type UDPConnWithTC struct {
	net.PacketConn
	OpFlag   int
	TxBucket *ratelimit.Bucket
	RxBucket *ratelimit.Bucket
}

func (c *UDPConnWithTC) ReadFrom(b []byte) (int, net.Addr, error) {
	n, a, err := c.PacketConn.ReadFrom(b)
	if err != nil {
		return n, a, err
	}
	if c.OpFlag&OpRateLimit != 0 && c.RxBucket != nil {
		c.RxBucket.Wait(int64(n * 8))
	}
	return n, a, err
}

func (c *UDPConnWithTC) WriteTo(b []byte, addr net.Addr) (int, error) {
	if c.OpFlag&OpRateLimit != 0 && c.TxBucket != nil {
		c.TxBucket.Wait(int64(len(b) * 8))
	}
	return c.PacketConn.WriteTo(b, addr)
}
