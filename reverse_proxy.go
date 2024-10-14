package main

import (
	"bufio"
	"net"
	"strings"
	// "strings"
	"log"
)

type container_conn struct {
	tcp net.Conn
	udp *net.UDPAddr
}

var ip_map = make(map[*net.IP]container_conn)

func ForwardTCP() error {
	// start tcp and upd listeners
	tcp_addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:9000")
	if err != nil {
		log.Println("[Server] failed to resolve address")
		panic(err)
	}
	tcp_listener, err := net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		log.Printf("[Server] failed to listen on %s", tcp_addr)
		panic(err)
	}

	for {
		conn, err := tcp_listener.Accept()
		if err != nil {
			log.Println("[Client] failed to accept connection")
			log.Println(err)
			return err
		}
		go func(conn net.Conn) {
			for {
				data, err := bufio.NewReader(conn).ReadBytes(byte(0))
				if err != nil {
					log.Println(err)
					return
				}
				ip := net.ParseIP(strings.Split(conn.RemoteAddr().String(), ":")[0])
				ip_map[&ip].tcp.Write(data)
			}
		}(conn)
	}

}

func ForwardUDP() error {
	udp_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		panic(err)
	}

	for {
		var buf [1024]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			// fmt.Println(err)
			return err
		}
		conn.WriteToUDP(buf[:0], ip_map[&addr.IP].udp)
	}
}

func AddForwardRule(ip string, id string) {
	// parsed_ip := net.ParseIP(ip)
	// ip_map[&parsed_ip] =  // TODO: start conn with container
	log.Println("[Server] Adding forwarding rule")
}

func RemoveForwardRule(ip string) {
	log.Println("[Server] Removing forwarding rule")
}
