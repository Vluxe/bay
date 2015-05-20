package main

import (
	"fmt"

	"github.com/Vluxe/bay"
	"github.com/fsouza/go-dockerclient"
)

type a int

func main() {
	var b a
	config := &bay.Config{
		CPU:            1,
		Memory:         50e6,
		DockerUrl:      "unix:///var/run/docker.sock",
		BuildInterface: b,
	}

	bay.Start(":8080", config)
}

func (aa a) PreBuild(buildDir, lang string, err error) {
}

func (aa a) PostBuild(container *docker.Container, lang string, err error) {
	if err != nil {
		fmt.Println(err)
	}

	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		fmt.Println(err)
	}

	m := map[docker.Port][]docker.PortBinding{
		"8080": []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: "8082"}},
	}

	fmt.Println(m)
	hostConfig := &docker.HostConfig{PortBindings: m}

	if err := client.StartContainer(container.ID, hostConfig); err != nil {
		fmt.Println(err)
	}
}
