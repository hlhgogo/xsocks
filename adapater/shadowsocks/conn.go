package shadowsocks

import (
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"io"
	"log"
	"net"
	"time"
)

const (
	udpBufSize     = 4 * 1024
	defaultTimeout = "60s"
)

// relay copies between left and right bidirectionally. Returns number of
// bytes copied from right to left, from left to right, and any error occurred.
func relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		var n int64
		var err error
		// n, err = copyAndPeep(right, left, cp.ClientWrite)
		n, err = io.Copy(right, left)
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()
	var n int64
	var err error
	// n, err = copyAndPeep(left, right, cp.ServerWrite)
	n, err = io.Copy(left, right)
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}
	return n, rs.N, err
}

func handleConn(conn net.Conn, opts ...SSOptionHandler) error {
	defer conn.Close()

	opt := &SSOption{}
	for _, o := range opts {
		o(opt)
	}

	tgt, err := socks.ReadAddr(conn)
	if err != nil {
		return err
	}
	//domain, _, _ := net.SplitHostPort(tgt.String())

	var rc net.Conn
	rc, err = net.Dial("tcp", tgt.String())
	if err != nil {
		log.Printf("failed to connect to target[%s]: %v", tgt.String(), err)
		return err
	}

	defer rc.Close()
	// rc.SetKeepAlive(true)

	var remoteConn net.Conn

	if opt.EnableTrafficControl && opt.RxBucket != nil && opt.TxBucket != nil {
		remoteConn = &ConnCustom{
			Conn:     rc,
			OpFlag:   OpRateLimit,
			TxBucket: opt.TxBucket,
			RxBucket: opt.RxBucket,
		}
	} else {
		remoteConn = rc
	}

	_, _, err = relay(conn, remoteConn)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil // ignore i/o timeout
		}
		log.Printf("relay error: %v", err)
		return err
	}
	return nil
}
