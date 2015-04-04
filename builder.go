package bay

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/samalba/dockerclient"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

const (
	golangEnv = iota // our environment types.
	rubyEnv
	pythonEnv
)

func BuildWithFiles(environment int, files [][]byte, fileNames []string) (string, error) {
	client, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return "", err
	}
	directoryName := randomStringName()
	os.Mkdir(directoryName, os.ModePerm)
	if len(fileNames) != len(files) {
		return "", errors.New("mismatch file bytes and name.")
	}
	for i, file := range files {
		f, err := os.Create(directoryName + "/" + fileNames[i])
		if err != nil {
			return "", err
		}
		if _, err = f.Write(file); err != nil { // write the file (bytes we got from the upload).
			return "", err
		}
	}

	dockerConfig, err := getDockerConfig(environment)
	if err != nil {
		return "", err
	}
	containerId, err := client.CreateContainer(dockerConfig, "something cool") // need a better name.
	if err != nil {
		return "", err
	}

	return containerId, nil
}

// randomStringName generates a random 32 character string.
func randomStringName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randomPortNumber picks a port between 30,000 to 31,000. Going forward we will need a way of picking ports deterministically.
func randomPortNumber() map[string]struct{} {
	rand.Seed(time.Now().Unix()) // not sure we need to see again, but w/e.
	min := 31000
	max := 30000
	var s struct{}
	ports := make(map[string]struct{})
	randomPort := fmt.Sprintf("%d", rand.Intn(max-min)+min)
	ports[randomPort] = s // not sure what the struct is for.
	return ports
}

// getDockerConfig returns a docker config depending on what environment/language chosen.
func getDockerConfig(environment int) (*dockerclient.ContainerConfig, error) {
	buildImage := ""
	if environment == golangEnv {
		buildImage = "golang:onbuild"
	}

	if buildImage == "" {
		return nil, errors.New("not a valid environment.")
	}

	return &dockerclient.ContainerConfig{Image: buildImage, Cmd: []string{"bash"}, ExposedPorts: randomPortNumber()}, nil
}
