package main

import (
	// "bufio"
	"io"
	"net"
	"sync"

	// "strings"

	// "strings"
	"log"
)

type container_conn struct {
	tcp net.Conn
	udp *net.UDPAddr
}

var ip_map = make(map[*net.IP]container_conn)

// func ForwardTCP() error {
// 	// start tcp and upd listeners
// 	tcp_addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:9000")
// 	if err != nil {
// 		log.Println("[Server] failed to resolve address")
// 		panic(err)
// 	}
// 	tcp_listener, err := net.ListenTCP("tcp", tcp_addr)
// 	if err != nil {
// 		log.Printf("[Server] failed to listen on %s", tcp_addr)
// 		panic(err)
// 	}
//
// 	for {
// 		conn, err := tcp_listener.Accept()
// 		if err != nil {
// 			log.Println("[Client] failed to accept connection")
// 			log.Println(err)
// 			return err
// 		}
// 		go func(conn net.Conn) {
// 			for {
// 				data, err := bufio.NewReader(conn).ReadBytes(byte(0))
// 				if err != nil {
// 					log.Println(err)
// 					return
// 				}
// 				ip := net.ParseIP(strings.Split(conn.RemoteAddr().String(), ":")[0])
// 				ip_map[&ip].tcp.Write(data)
// 			}
// 		}(conn)
// 	}
//
// }
//
// func ForwardUDP() error {
// 	udp_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	conn, err := net.ListenUDP("udp", udp_addr)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	for {
// 		var buf [1024]byte
// 		_, addr, err := conn.ReadFromUDP(buf[0:])
// 		if err != nil {
// 			// fmt.Println(err)
// 			return err
// 		}
// 		conn.WriteToUDP(buf[:0], ip_map[&addr.IP].udp)
// 	}
// }

func AddForwardRule(ip string, id string) {
	// parsed_ip := net.ParseIP(ip)
	// ip_map[&parsed_ip] =  // TODO: start conn with container
	log.Println("[Server] Adding forwarding rule")
}

func RemoveForwardRule(ip string) {
	log.Println("[Server] Removing forwarding rule")
}

type Proxy struct {
	addr  string
	rules sync.Map
}

func (p *Proxy) ServeListener(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("[Proxy] failed to accept connection")
			log.Println(err)
		}

		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(c net.Conn) {
	remoteAddr := c.RemoteAddr

	srvAddr, _ := p.rules.Load(remoteAddr)

	gameServerConn, err := net.Dial("tcp", srvAddr.(string))
	if err != nil {
		log.Println("[Proxy] failed to connect to container")
		log.Println(err)
		return
	}

	// forward from client to server
	go func() {
		io.Copy(gameServerConn, c)

		gameServerConn.Close()
	}()

	// forward from server to client
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := gameServerConn.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				c.Write(buf[:n])
			}
		}
	}()
}

func (p *Proxy) addForwardRule(src string, dst string) {
	p.rules.Store(src, dst)
}

func (p *Proxy) removeForwardRule(src string) {
	p.rules.Delete(src)
}
