package shadowsocks

import (
	"context"
	"fmt"
	"github.com/juju/ratelimit"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"log"
	"net"
)

type ss struct {
	Port           int    `json:"server_port"`
	Method         string `json:"method"`
	Password       string `json:"password"`
	tcpRunning     bool
	udpCtx         context.Context
	udpCancel      context.CancelFunc
	txBucket       *ratelimit.Bucket
	rxBucket       *ratelimit.Bucket
	listener       net.Listener
	udpPackageConn net.PacketConn
	isRateLimit    bool
	Stat           *Stat
}

type Stat struct {
	TcpSendBytes   uint64
	TcpRecvBytes   uint64
	UdpSendBytes   uint64
	UdpRecvBytes   uint64
	UdpSendPackets uint64
	UdpRecvPackets uint64
}

func (s *ss) SetRateLimit(tRate, rRate, tCapacity, rCapacity int64) {
	s.txBucket = ratelimit.NewBucketWithRate(float64(tRate), tCapacity)
	s.rxBucket = ratelimit.NewBucketWithRate(float64(rRate), rCapacity)
}

func NewSS(port int, method, password string) *ss {
	udpCtx, udpCancel := context.WithCancel(context.Background())
	return &ss{
		Port:        port,
		Method:      method,
		Password:    password,
		isRateLimit: false,
		udpCtx:      udpCtx,
		udpCancel:   udpCancel,
		Stat:        &Stat{},
	}
}

func (s *ss) RunTcp() error {
	ciph, err := core.PickCipher(s.Method, []byte{}, s.Password)
	if err != nil {
		log.Printf("Create SS on port [%d] failed: %s", s.Port, err)
		return err
	}

	addr := fmt.Sprintf(":%d", s.Port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen on %s : %v", addr, err)
		return err
	}
	s.tcpRunning = true

	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				log.Printf("failed to accept: %v", err)
				continue
			}
			c.(*net.TCPConn).SetKeepAlive(true)
			go func() {
				c = ciph.StreamConn(c)

				opts := []SSOptionHandler{}
				if s.isRateLimit && s.txBucket != nil && s.rxBucket != nil {
					opts = append(opts, withTrafficControl(s.txBucket, s.rxBucket))
				}

				err = handleConn(c, opts...)

			}()
		}
	}()

	return nil
}

func (s *ss) RunUDP() error {

	var err error
	ciph, err := core.PickCipher(s.Method, []byte{}, s.Password)
	if err != nil {
		log.Printf("Create SS on port [%d] failed: %s", s.Port, err)
		return err
	}

	addr := fmt.Sprintf(":%d", s.Port)
	s.udpPackageConn, err = net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("ListenPacket error: %s", err)
	}
	s.udpPackageConn = ciph.PacketConn(s.udpPackageConn)
	localAddr := "0.0.0.0"

	s.udpPackageConn = WrapPacketConn(s.udpPackageConn, s.Stat)
	s.udpPackageConn = ciph.PacketConn(s.udpPackageConn)

	opts := []SSOptionHandler{}
	if s.isRateLimit && s.txBucket != nil && s.rxBucket != nil {
		opts = append(opts, withTrafficControl(s.txBucket, s.rxBucket))
	}

	go handlePacketConn(s.udpCtx, s.udpPackageConn, localAddr, opts...)

	return nil
}
