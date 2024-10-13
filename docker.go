package main

import (
	"errors"
	"io"
	"net"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var cli *client.Client

const IMAGE = "alpine"    // TODO: get from env var
const MAX_CONTAINERS = 10 // TODO: get from env var

var num_containers = 0

var container_map = make(map[string]net.IP)

func StartContainer() (string, error) {
	if num_containers >= MAX_CONTAINERS {
		return "", errors.New("max containers")
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"sleep", "30"},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		// fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to create container")
		// fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		// fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to start container")
		// fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	json, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}

	container_map[resp.ID] = net.ParseIP(json.NetworkSettings.IPAddress)

	num_containers += 1
	return resp.ID, nil
}

func StopContainer(id string) error {
	err := cli.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		// fmt.Fprintln(os.Stderr, "[Server] ERROR: failed to stop running container")
		// fmt.Fprintln(os.Stderr, err)
		return err
	}
	// fmt.Println("[Server] Stopped running container")
	num_containers -= 1
	delete(container_map, id)

	return nil
}

func ConnectDocker() error {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// pull image
	out, err := cli.ImagePull(ctx, IMAGE, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	return nil
}

func ContainerHealthCheck() {

}

func CleanupDocker() error {
	for id := range container_map {
		StopContainer(id)
	}

	cli.Close()

	return nil
}
