package main

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestReverseProxyTcp(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	proxy := Proxy{}
	proxy.AddForwardRule("127.0.0.1:6001", "127.0.0.1:6002")
	go func() {
		proxy.Start(ctx, "127.0.0.1:6000")
	}()

	c := make(chan bool)

	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:6002")
		if err != nil {
			t.Error(err)
		}
		defer l.Close()
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		buffer := make([]byte, 11)
		_, err = conn.Read(buffer)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(buffer, []byte("hello world")) {
			t.Errorf("buffer == \"hello world\" failed, got \"%s\"", buffer)
		}
		c <- true
	}()

	time.Sleep(1 * time.Second)

	lAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:6001")
	rAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:6000")
	conn, err := net.DialTCP("tcp", lAddr, rAddr)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		t.Error(err)
	}
	_, err = conn.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}
	<-c
	proxy.RemoveForwardRule("127.0.0.1:6001")
	cancelFunc()

	time.Sleep(4 * time.Second)
}

func TestReverseProxyUdp(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	proxy := Proxy{}
	proxy.AddForwardRule("127.0.0.1:7001", "127.0.0.1:7002")
	go func() {
		proxy.Start(ctx, "127.0.0.1:7000")
	}()

	c := make(chan bool)

	go func() {
		lAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7002")
		conn, err := net.ListenUDP("udp", lAddr)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		buffer := make([]byte, 11)
		conn.SetDeadline(time.Now().Add(10 * time.Second))
		_, _, _ = conn.ReadFromUDP(buffer[0:])
		if !bytes.Equal(buffer, []byte("hello world")) {
			t.Errorf("buffer == \"hello world\" failed, got %s", buffer)
		}
		c <- true
	}()

	time.Sleep(1 * time.Second)

	lAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7001")
	rAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7000")
	conn, err := net.DialUDP("udp", lAddr, rAddr)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	conn.Write([]byte("hello world"))
	<-c
	proxy.RemoveForwardRule("127.0.0.1:7001")
	cancelFunc()

	time.Sleep(4 * time.Second)
}
