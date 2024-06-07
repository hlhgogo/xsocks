package shadowsocks

import (
	"context"
	"fmt"
	"github.com/shadowsocks/go-shadowsocks2/core"
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

func handleConn(conn net.Conn) error {
	defer conn.Close()
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
	remoteConn = rc

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

func runTcp(method, password string, port int) error {

	ciph, err := core.PickCipher(method, []byte{}, password)
	if err != nil {
		log.Printf("Create SS on port [%d] failed: %s", port, err)
		return err
	}

	addr := fmt.Sprintf(":%d", port)
	tcpConn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen on %s : %v", addr, err)
		return err
	}

	go func() {
		for {
			c, err := tcpConn.Accept()
			if err != nil {
				log.Printf("failed to accept: %v", err)
				continue
			}
			c.(*net.TCPConn).SetKeepAlive(true)
			go func() {
				c = ciph.StreamConn(c)
				err = handleConn(c)
			}()
		}
	}()

	//log.Printf("listening TCP on %s", addr)
	return nil
}

func runUDP(method, password string, port int) error {
	var err error
	ciph, err := core.PickCipher(method, []byte{}, password)
	if err != nil {
		log.Printf("Create SS on port [%d] failed: %s", port, err)
		return err
	}
	addr := fmt.Sprintf(":%d", port)
	udpPackageConn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Printf("UDP remote listen error: %v", err)
		return err
	}
	udpPackageConn = ciph.PacketConn(udpPackageConn)
	localAddr := "0.0.0.0"

	go handlePacketConn(context.TODO(), udpPackageConn, localAddr)

	return nil
}

func StartSs(method, password string, port int) {

	err := runTcp(method, password, port)
	if err != nil {
		panic(err)
	}

	err = runUDP(method, password, port)
	if err != nil {
		panic(err)
	}

	log.Printf("Start shadowsocks [%s|%s|%d]", method, password, port)
	select {}
}
