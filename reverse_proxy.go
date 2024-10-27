package main

import (
	"fmt"
	"io"
	"net"
	"sync"

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

func (p *Proxy) Start(srcAddr string) {
	go p.startTcpListener(srcAddr)
	p.startUdpListener(srcAddr)
}

func (p *Proxy) GetUdpConn(dstAddr string) (*net.UDPConn, error) {
	if conn, ok := p.udpConn.Load(dstAddr); ok {
		return conn.(*net.UDPConn), nil
	}

	p.connMutex.Lock()
	defer p.connMutex.Unlock()

	if conn, ok := p.udpConn.Load(dstAddr); ok {
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

func (p *Proxy) startTcpListener(srcAddr string) {
	listener, err := net.Listen("tcp", srcAddr)
	if err != nil {
		log.Println("[Proxy] tcp listener error")
		log.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[Proxy] error accepting tcp connection")
			log.Println(err)
			continue
		}
		go p.handleTcpConnection(conn, srcAddr)
	}
}

func (p *Proxy) startUdpListener(srcAddr string) {
	conn, err := net.ListenPacket("udp", srcAddr)
	if err != nil {
		log.Println("[Proxy] failed to listen on udp")
		log.Println(err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1500) // 1500 is default MTU

	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println("[Proxy] error reading udp packet") // can probably not log this
			continue
		}

		go p.handleUdpConnection(buffer[:n], addr, srcAddr)
	}
}

func (p *Proxy) handleTcpConnection(c net.Conn, srcAddr string) {
	defer c.Close()

	dstAddr, ok := p.rules.Load(srcAddr)
	if !ok {
		log.Printf("[Proxy] No rule found for %s\n", srcAddr)
		return
	}

	dstConn, err := net.Dial("tcp", dstAddr.(string))
	if err != nil {
		log.Println("[Proxy] failed to connect to container")
		log.Println(err)
		return
	}
	defer dstConn.Close()

	go io.Copy(dstConn, c)
	io.Copy(c, dstConn)
}

func (p *Proxy) handleUdpConnection(buffer []byte, srcAddr net.Addr, ruleSrcAddr string) {
	dstAddr, ok := p.rules.Load(ruleSrcAddr)
	if !ok {
		log.Printf("[Proxy] no rule found for %s\n", ruleSrcAddr)
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
