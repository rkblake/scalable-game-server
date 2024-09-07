package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const IMAGE = "alpine"
const MAX_CONTAINERS = 10

var cli *client.Client
var ctx context.Context
var num_containers int = 0

var tcp_map map[string]net.Conn

func start_container(ip string) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: IMAGE}, nil, nil, nil, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to create container")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to start container")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println("[Server] Started container")
	num_containers += 1

	// add container conn to map
	tcp_addr, err := net.ResolveTCPAddr("tcp4", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to establish tcp connection to container")
		fmt.Fprintln(os.Stderr, err)
		stop_container(resp.ID)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcp_addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to establish tcp connection to container")
		fmt.Fprintln(os.Stderr, err)
		stop_container(resp.ID)
		return
	}

	go func(id string, conn net.Conn) {

		buf := make([]byte, 16)
		conn.SetDeadline(time.Now().Add(time.Minute))
		for true {
			_, err = conn.Write([]byte("keepalive"))
			if err != nil {
				fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to send TCP to container")
				fmt.Fprintln(os.Stderr, err)
				stop_container(id)
				return
			}

			buf = []byte("")

			_, err = conn.Read(buf)
			if err != nil {
				fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to receive TCP from container")
				fmt.Fprintln(os.Stderr, err)
				stop_container(id)
				return
			}

			if string(buf) == "keepalive" {
				continue
			} else if string(buf) == "stop" {
				fmt.Println("[Server] Received stop from container; stopping container")
				stop_container(id)
				return
			} else {
				fmt.Println("[Server] Did not receive valid response from container")
				stop_container(id)
				return
			}
		}

	}(resp.ID, conn)
}

func stop_container(id string) {
	err := cli.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to stop running container")
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("[Server] Stopped running container")
	num_containers -= 1
}

func join_container(ip string) {}

func get_create_match(w http.ResponseWriter, r *http.Request) {
	if num_containers >= MAX_CONTAINERS {
		return // at max capacity
		// TODO: send response
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	start_container(ip)
}
func get_join_match(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	join_container(ip)
}

func tcp_worker() {}
func udp_worker() {}

func main() {
	fmt.Println("[Server] Starting server...")
	ctx := context.Background()

	// connect to docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// pull image
	out, err := cli.ImagePull(ctx, IMAGE, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	// io.Copy(os.Stdout, out)

	// setup http endpoints
	http.HandleFunc("/create-match", get_create_match)
	http.HandleFunc("/join-match", get_join_match)

	// start tcp and upd listeners
	tcp_addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	tcp_listener, err := net.ListenTCP("tcp", tcp_addr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := tcp_listener.Accept()
			if err != nil {
				fmt.Println(err)
			}
			go func(conn net.Conn) {
				for {
					data, err := bufio.NewReader(conn).ReadBytes(byte(0))
					if err != nil {
						fmt.Println(err)
						return
					}
					tcp_map[conn.RemoteAddr().String()].Write(data)
				}
			}(conn)
		}
	}()

	go func() {
		for {
			udp_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
			if err != nil {
				fmt.Println(err)
				return
			}

			conn, err := net.ListenUDP("udp", udp_addr)
			if err != nil {
				fmt.Println(err)
				return
			}
			for {
				var buf [1024]byte
				_, addr, err := conn.ReadFromUDP(buf[0:])
				if err != nil {
					fmt.Println(err)
					return
				}
				tcp_map[addr.String()].Write(buf[0:])
			}
		}
	}()

	fmt.Println("[Server] Started")

	// start http server
	http.ListenAndServe(":3333", nil)

}
