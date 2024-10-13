package main

import (
	"bufio"
	"net"
)

var tcp_map map[*net.Conn]net.Conn
var udp_map map[*net.UDPAddr]net.Conn
var ip_map map[*net.IP]net.IP

func ForwardTCP() error {
	// start tcp and upd listeners
	tcp_addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	tcp_listener, err := net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := tcp_listener.Accept()
		if err != nil {
			// fmt.Println(err)
			return err
		}
		go func(conn net.Conn) {
			for {
				data, err := bufio.NewReader(conn).ReadBytes(byte(0))
				if err != nil {
					// fmt.Println(err)
					return
				}
				tcp_map[&conn].Write(data)
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
		// tcp_map[addr.String()].Write(buf[0:])
		udp_map[addr].Write(buf[0:])
	}

	return nil
}

func AddForwardRule(ip string, id string) {
	parsed_ip := net.ParseIP(ip)
	ip_map[&parsed_ip] = container_map[id]
}

func RemoveForwardRule(ip string) {

}
