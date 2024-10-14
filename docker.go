package main

import (
	"bytes"
	"errors"
	// "fmt"

	// "io"
	"log"
	"math"
	"net"

	// "os"
	"time"

	"github.com/docker/docker/api/types/container"
	// "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var cli *client.Client

const IMAGE = "server"    // TODO: get from env var
const MAX_CONTAINERS = 10 // TODO: get from env var

var num_containers = 0

type Status int

const (
	lobby Status = iota
	in_game
	finished
)

type container_data struct {
	ip          net.IP
	num_players int
	max_players int
	private     bool
	in_progress bool
	status      Status
	conn        net.Conn
}

var container_map = make(map[string]container_data)

func StartContainer(max_players int, private bool) (string, error) {
	log.Println("[Server] starting container")
	if num_containers >= MAX_CONTAINERS {
		log.Println("[Server] Failed to start container: at max capacity. ")
		return "", errors.New("max containers")
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: IMAGE,
		// Cmd:   []string{"sleep", "30"},
		// Tty:   false,
	}, &container.HostConfig{
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		log.Println("[Server] ERROR: failed to create container")
		log.Println(err)
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Println("[Server] ERROR: failed to start container")
		log.Println(err)
		return "", err
	}

	// get containers ip address
	json, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}
	ip := net.ParseIP(json.NetworkSettings.IPAddress)

	// TODO: container should be the server and container should call back when ready
	time.Sleep(5 * time.Second)
	conn, err := net.DialTimeout("tcp", ip.String()+":9001", 30*time.Second)
	if err != nil {
		log.Println("[Container] failed to connect to container")
		log.Println(err)
		StopContainer(resp.ID)
		return "", err
	}

	container_map[resp.ID] = container_data{
		ip:          ip,
		num_players: 1,
		max_players: max_players,
		private:     private,
		in_progress: false,
		conn:        conn,
	}

	num_containers += 1
	return resp.ID, nil
}

func StopContainer(id string) error {
	err := cli.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Println("[Server] ERROR: failed to stop running container")
		log.Println(err)
		return err
	}
	log.Println("[Server] Stopped running container")
	num_containers -= 1
	delete(container_map, id)

	return nil
}

func ConnectDocker() error {
	log.Println("[Server] connecting to docker")
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// TODO: check if image exists
	// pull image
	// out, err := cli.ImagePull(ctx, IMAGE, image.PullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// defer out.Close()
	// io.Copy(os.Stdout, out)

	log.Println("[Server] successfuly connected to docker")
	return nil
}

// TODO: change to starting a goroutine for each container, maybe create a queue of
// containers that neeed stopping
func ContainerHealthCheck() {
	go func() {
		start := time.Now()
		for {
			elapsed := time.Since(start)
			max := math.Max((1*time.Minute - elapsed).Seconds(), 0)
			time.Sleep(time.Duration(max * float64(time.Second)))

			// log.Println("[Server] running healthcheck") // maybe remove? cluttering logs
			start = time.Now()
			for k, v := range container_map {
				v.conn.Write([]byte("status?\n"))

				reply := make([]byte, 16)

				v.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				_, err := v.conn.Read(reply)
				if err != nil {
					log.Println("[Container] failed to receive packet from container")
					StopContainer(k)
					RemoveMatch(k)
					continue
				}
				if bytes.Equal(reply, []byte("lobby")) {
					v.status = lobby
				} else if bytes.Equal(reply, []byte("game")) {
					v.status = in_game
				} else if bytes.Equal(reply, []byte("finished")) {
					log.Println("[Container] returned status: finished")
					StopContainer(k)
					RemoveMatch(k)
					continue
				}
			}
		}
	}()
}

func CleanupDocker() error {
	for id := range container_map {
		StopContainer(id)
	}

	cli.Close()

	return nil
}
