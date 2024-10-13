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
	ctx = context.Background()

	// connect to docker
	err := ConnectDocker()
	if err != nil {
		panic(err)
	}
	_, err = StartContainer()
	if err != nil {
		panic(err)
	}

	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	panic(err)
	// }
	// defer cli.Close()
	//
	// reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", image.PullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// io.Copy(os.Stdout, reader)
	//
	// resp, err := cli.ContainerCreate(ctx, &container.Config{
	// 	Image: "alpine",
	// 	Cmd:   []string{"echo", "hello world"},
	// }, nil, nil, nil, "")
	// if err != nil {
	// 	panic(err)
	// }
	//
	// if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
	// 	panic(err)
	// }
	//
	// out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	// if err != nil {
	// 	panic(err)
	// }
	//
	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	// setup http endpoints
	// http.HandleFunc("/create-match", get_create_match)
	// http.HandleFunc("/join-match", get_join_match)

	fmt.Println("[Server] Started")

	// start http server
	// http.ListenAndServe(":3333", nil)
	HandleEndpoints()

}
