package bay

import (
	"archive/tar"
	"bytes"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/libgit2/git2go"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

//const (
//	GolangEnv = iota // our environment types.
//	rubyEnv
//	railsEnv
//	pythonEnv
//	djangoEnv
//)

type tarBuilder struct {
	tr      *tar.Writer
	rootDir string
}

// BuildWithGitRepo takes a git url to build a docker image.
// Returns the container ID of the newly built docker image.
func BuildWithGitRepo(gitUrl string) (string, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock") // probably need to be configurable.
	if err != nil {
		return "", err
	}
	directoryName := randomStringName()
	os.Mkdir(directoryName, os.ModePerm)

	if _, err := git.Clone(gitUrl, directoryName, &git.CloneOptions{}); err != nil {
		os.RemoveAll(directoryName)
		return "", err
	}

	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	tb := tarBuilder{tr: tar.NewWriter(inputbuf), rootDir: directoryName}
	if err := filepath.Walk(directoryName, tb.pathWalker); err != nil {
		os.RemoveAll(directoryName)
		return "", err
	}

	tb.tr.Flush()
	tb.tr.Close()
	opts := docker.BuildImageOptions{
		Name:         "name",
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}
	if err := client.BuildImage(opts); err != nil {
		os.RemoveAll(directoryName)
		return "", err
	}

	os.RemoveAll(directoryName)
	return "", nil
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

// pathWalker is the helper function used for filepath Walk.
func (tb *tarBuilder) pathWalker(path string, info os.FileInfo, err error) error {
	if info.IsDir() || strings.HasPrefix(".git", path) {
		return nil // skip this one.
	}

	name, err := filepath.Rel(tb.rootDir, path)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	now := time.Now()
	header := &tar.Header{Name: name, Size: info.Size(), Mode: int64(info.Mode()),
		ModTime: info.ModTime(), AccessTime: now, ChangeTime: now,
	}
	if err := tb.tr.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tb.tr, file); err != nil {
		return err
	}
	return nil
}

// randomPortNumber picks a port between 30,000 to 31,000. Going forward we will need a way of picking ports deterministically.
//func randomPortNumber() map[string]struct{} {
//	rand.Seed(time.Now().Unix()) // not sure we need to see again, but w/e.
//	min := 31000
//	max := 30000
//	var s struct{}
//	ports := make(map[string]struct{})
//	randomPort := fmt.Sprintf("%d", rand.Intn(max-min)+min)
//	ports[randomPort] = s // not sure what the struct is for.
//	return ports
//}

// getDockerConfig returns a docker config depending on what environment/language chosen.
//func getDockerConfig(environment int) (*dockerclient.ContainerConfig, error) {
//	buildImage := ""
//	switch environment {
//	case GolangEnv:
//		buildImage = "golang:onbuild"
//	case rubyEnv:
//		buildImage = "rails:onbuild"
//	case railsEnv:
//		buildImage = "ruby:onbuild"
//	case pythonEnv:
//		buildImage = "python:onbuild"
//	case djangoEnv:
//		buildImage = "django:onbuild"
//	default:
//		return nil, errors.New("not a valid environment.")
//	}

//	return &dockerclient.ContainerConfig{Image: buildImage, Cmd: []string{"bash"}, ExposedPorts: randomPortNumber()}, nil
//}

// BuildWithFiles takes files and an environment to build a docker image.
// Returns the container ID of the newly built docker image.
//func BuildWithFiles(environment int, files [][]byte, fileNames []string) (string, error) {
//	client, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil) // probably need to be configurable.
//	if err != nil {
//		return "", err
//	}
//	directoryName := randomStringName()
//	os.Mkdir(directoryName, os.ModePerm)
//	if len(fileNames) != len(files) {
//		return "", errors.New("mismatch file bytes and name.")
//	}
//	for i, file := range files {
//		f, err := os.Create(directoryName + "/" + fileNames[i])
//		if err != nil {
//			return "", err
//		}
//		if _, err = f.Write(file); err != nil { // write the file (bytes we got from the upload).
//			return "", err
//		}
//	}
//
//	dockerConfig, err := getDockerConfig(environment)
//	if err != nil {
//		return "", err
//	}
//	containerId, err := client.CreateContainer(dockerConfig, "something cool") // need a better name.
//	if err != nil {
//		return "", err
//	}
//
//	return containerId, nil
//}
