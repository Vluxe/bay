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
	fmt.Println(err)
}

func (aa a) PostBuild(container *docker.Container, lang string, err error) {
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(container.ID)

	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(container.Image)
	fmt.Println(container.Name)
	if err := client.StartContainer(container.ID, container.HostConfig); err != nil {
		fmt.Println(err)
	}
}
