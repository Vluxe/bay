package bay

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/libgit2/git2go"
	"github.com/samalba/dockerclient"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

const (
	GolangEnv = iota // our environment types.
	RubyEnv
	RailsEnv
	PythonEnv
	DjangoEnv
)

// BuildWithGitRepo takes a git url and an environment to build a docker image.
// Returns the container ID of the newly built docker image.
func BuildWithGitRepo(environment int, gitUrl string, gitCommit string) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock") // probably need to be configurable.
	if err != nil {
		return "", err
	}
	directoryName := gitCommit
	os.Mkdir(directoryName, os.ModePerm)

	if _, err := git.Clone(gitUrl, directoryName, &git.CloneOptions{}); err != nil {
		return "", nil
	}

	f, err := os.Open(directoryName + "/Dockerfile")
	if err != nil {
		return "", err
	}
	t := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: 10, ModTime: t, AccessTime: t, ChangeTime: t})
	io.Copy(tr, f)
	tr.Close()
	opts := docker.BuildImageOptions{
		Name:         "test",
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}
	if err := client.BuildImage(opts); err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	io.Copy(buffer, outputbuf)         // get rid of this
	fmt.Println("buffer is: ", buffer) // debugging crap.
	return "", nil
}

// BuildWithFiles takes files and an environment to build a docker image.
// Returns the container ID of the newly built docker image.
func BuildWithFiles(environment int, files [][]byte, fileNames []string) (string, error) {
	client, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil) // probably need to be configurable.
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
	switch environment {
	case GolangEnv:
		buildImage = "golang:onbuild"
	case RubyEnv:
		buildImage = "rails:onbuild"
	case RailsEnv:
		buildImage = "ruby:onbuild"
	case PythonEnv:
		buildImage = "python:onbuild"
	case DjangoEnv:
		buildImage = "django:onbuild"
	default:
		return nil, errors.New("not a valid environment.")
	}

	return &dockerclient.ContainerConfig{Image: buildImage, Cmd: []string{"bash"}, ExposedPorts: randomPortNumber()}, nil
}
