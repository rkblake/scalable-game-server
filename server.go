package main

import (
	"context"
	"fmt"
)

var ctx = context.Background()

// 	// add container conn to map
// 	tcp_addr, err := net.ResolveTCPAddr("tcp4", "")
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to establish tcp connection to container")
// 		fmt.Fprintln(os.Stderr, err)
// 		StopContainer(id)
// 		return
// 	}
//
// 	conn, err := net.DialTCP("tcp", nil, tcp_addr)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to establish tcp connection to container")
// 		fmt.Fprintln(os.Stderr, err)
// 		StopContainer(id)
// 		return
// 	}
//
// 	go func(id string, conn net.Conn) {
//
// 		buf := make([]byte, 16)
// 		conn.SetDeadline(time.Now().Add(time.Minute))
// 		for true {
// 			_, err = conn.Write([]byte("keepalive"))
// 			if err != nil {
// 				fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to send TCP to container")
// 				fmt.Fprintln(os.Stderr, err)
// 				StopContainer(id)
// 				return
// 			}
//
// 			buf = []byte("")
//
// 			_, err = conn.Read(buf)
// 			if err != nil {
// 				fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to receive TCP from container")
// 				fmt.Fprintln(os.Stderr, err)
// 				StopContainer(id)
// 				return
// 			}
//
// 			if string(buf) == "keepalive" {
// 				continue
// 			} else if string(buf) == "stop" {
// 				fmt.Println("[Server] Received stop from container; stopping container")
// 				StopContainer(id)
// 				return
// 			} else {
// 				fmt.Println("[Server] Did not receive valid response from container")
// 				StopContainer(id)
// 				return
// 			}
// 		}
//
// 	}(id, conn)
// }

func main() {
	fmt.Println("[Server] Starting server...")

	// connect to docker
	err := ConnectDocker()
	if err != nil {
		panic(err)
	}

	fmt.Println("[Server] Started")

	// start http server
	HandleEndpoints()

}
