package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"log"
)

type Proxy struct {
	rules     sync.Map
	udpConn   sync.Map
	connMutex sync.Mutex
}

func (p *Proxy) AddForwardRule(src string, dst string) {
	p.rules.Store(src, dst)
	log.Printf("[Proxy] Adding forwarding rule: %s -> %s\n", src, dst)
}

func (p *Proxy) RemoveForwardRule(src string) {
	p.rules.Delete(src)
	log.Printf("[Proxy] Removing forwarding rule for: %s\n", src)
}

func (p *Proxy) Start(ctx context.Context, srcAddr string) {
	go p.startTcpListener(ctx, srcAddr)
	p.startUdpListener(ctx, srcAddr)

	// cleanup
	p.connMutex.Lock()
	defer p.connMutex.Unlock()
	// m := map[string]interface{}{}
	p.udpConn.Range(func(key, value interface{}) bool {
		value.(*net.UDPConn).Close()
		return true
	})
}

func (p *Proxy) GetUdpConn(dstAddr string) (*net.UDPConn, error) {
	srcIp := strings.Split(dstAddr, ":")[0]
	if conn, ok := p.udpConn.Load(srcIp); ok {
		return conn.(*net.UDPConn), nil
	}

	p.connMutex.Lock()
	defer p.connMutex.Unlock()

	if conn, ok := p.udpConn.Load(srcIp); ok {
		return conn.(*net.UDPConn), nil
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", dstAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve udp addr %s: %v", dstAddr, err)
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to udp %s: %v", dstAddr, err)
	}

	p.udpConn.Store(dstAddr, conn)
	return conn, nil
}

func (p *Proxy) startTcpListener(ctx context.Context, srcAddr string) {
	lAddr, err := net.ResolveTCPAddr("tcp", srcAddr)
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		log.Println("[Proxy] tcp listener error")
		log.Println(err)
		return
	}
	defer listener.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			listener.SetDeadline(time.Now().Add(time.Second))
			conn, err := listener.Accept()
			if err != nil {
				if !err.(*net.OpError).Timeout() {
					log.Println("[Proxy] error accepting tcp connection")
					log.Println(err)
				}
				continue
			}
			go p.handleTcpConnection(conn)
		}
	}
}

func (p *Proxy) handleTcpConnection(src net.Conn) {
	defer src.Close()

	dstAddr, ok := p.rules.Load(strings.Split(src.RemoteAddr().String(), ":")[0])
	if !ok {
		log.Printf("[Proxy] No rule found for %s\n", src.RemoteAddr().String())
		return
	}

	dstConn, err := net.Dial("tcp", dstAddr.(string))
	if err != nil {
		log.Println("[Proxy] failed to connect to container")
		log.Println(err)
		return
	}
	defer dstConn.Close()

	go io.Copy(dstConn, src)
	io.Copy(src, dstConn)
}

func (p *Proxy) startUdpListener(ctx context.Context, srcAddr string) {
	conn, err := net.ListenPacket("udp", srcAddr)
	if err != nil {
		log.Println("[Proxy] failed to listen on udp")
		log.Println(err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1500) // 1500 is default MTU

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn.SetDeadline(time.Now().Add(time.Second))
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				if !err.(*net.OpError).Timeout() {
					log.Println("[Proxy] error reading udp packet") // can probably not log this
					log.Println(err)
				}
				continue
			}

			go p.handleUdpConnection(buffer[:n], addr)
		}
	}
}

func (p *Proxy) handleUdpConnection(buffer []byte, srcAddr net.Addr) {
	dstAddr, ok := p.rules.Load(strings.Split(srcAddr.String(), ":")[0])
	if !ok {
		log.Printf("[Proxy] no rule found for %s\n", srcAddr.String())
		return
	}

	conn, err := p.GetUdpConn(dstAddr.(string))
	if err != nil {
		log.Println("[Proxy] failed getting udp conn")
		log.Println(err)
		return
	}

	_, err = conn.Write(buffer)
	if err != nil {
		log.Println("[Proxy] failed to forward packets")
		log.Println(err)
		return
	}
}
