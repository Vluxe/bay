package bay

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/libgit2/git2go"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

type tarBuilder struct {
	tr      *tar.Writer
	rootDir string
}

// pathWalker is the helper function used for filepath Walk called in BuildWithGitRepo.
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

// buildWithGit takes a git url and language to build a docker image.
func (s *server) buildWithGit(gitUrl, lang string) {
	dir, err := ioutil.TempDir("", "bay-")
	if err != nil {
		s.callBack("", lang, err)
		return
	}
	defer os.RemoveAll(dir)

	if _, err := git.Clone(gitUrl, dir, &git.CloneOptions{}); err != nil {
		s.callBack(dir, lang, err)
		return
	}

	s.callBack(dir, lang, nil)
	s.createContainer(dir, lang)
}

// buildWithFiles takes either a source file or zip file, language to build a docker image.
func (s *server) buildWithFiles(f multipart.File, name, lang, contentType string) {
	defer f.Close()

	dir, err := ioutil.TempDir("", "bay-")
	if err != nil {
		s.callBack("", lang, err)
		return
	}
	defer os.RemoveAll(dir)

	tmpFile, err := os.Create(dir + "/" + name)
	if err != nil {
		s.callBack(dir, lang, err)
		return
	}

	_, err = io.Copy(tmpFile, f)
	if err != nil {
		s.callBack(dir, lang, err)
		return
	}

	if contentType == "application/zip" {
		fstat, err := tmpFile.Stat()
		if err != nil {
			s.callBack(dir, lang, err)
			return
		}
		r, err := zip.NewReader(f, fstat.Size())
		for _, zfile := range r.File {
			tmp, err := ioutil.TempFile(dir, "upload-")
			if err != nil {
				s.callBack(dir, lang, err)
				return
			}
			rc, err := zfile.Open()
			_, err = io.Copy(tmp, rc)
			if err != nil {
				s.callBack(dir, lang, err)
				return
			}
		}
	}

	s.callBack(dir, lang, nil)
	s.createContainer(dir, lang)
}

// createContainer is a simple factory function for doing the actual docker image building.
func (s *server) createContainer(dir, lang string) {
	// need to write a new Dockerfile in dir with the an image and port
	dockerFile, err := os.Create(dir + "/Dockerfile")
	if err != nil {
		s.callBack(dir, lang, err)
		return
	}
	contents := []byte(fmt.Sprintf("FROM %s\n EXPOSE %d", imageFromLang(lang), 8080))
	if _, err := dockerFile.Write(contents); err != nil {
		s.callBack(dir, lang, err)
		return
	}

	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	tb := tarBuilder{tr: tar.NewWriter(inputbuf), rootDir: dir}
	if err := filepath.Walk(dir, tb.pathWalker); err != nil {
		s.callBack(dir, lang, err)
		return
	}

	imageName := generateRandomImageName()
	tb.tr.Flush()
	tb.tr.Close()
	opts := docker.BuildImageOptions{
		Name:         imageName,
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}
	if err := s.dockerClient.BuildImage(opts); err != nil {
		s.callBack(dir, lang, err)
		return
	}

	config := &docker.Config{
		CPUShares:       s.config.CPU,
		Memory:          s.config.Memory,
		Tty:             true,
		OpenStdin:       false,
		Image:           imageName,
		NetworkDisabled: false,
		//ExposedPorts:    port,
	}

	hostConfig := &docker.HostConfig{} // set our container privileges...

	containerOpts := docker.CreateContainerOptions{Name: "", Config: config, HostConfig: hostConfig}
	container, err := s.dockerClient.CreateContainer(containerOpts)

	if s.config.BuildInterface != nil {
		s.config.BuildInterface.PostBuild(container, lang, err)
	}
}

// callBack wraps up the PreBuild call interface check into a clean little function.
func (s *server) callBack(dir, lang string, err error) {
	if s.config.BuildInterface != nil {
		s.config.BuildInterface.PreBuild(dir, lang, err)
	}
}

// imageFromLang figures out which docker to use based on the lang chosen.
func imageFromLang(lang string) string {
	switch lang {
	case "c":
		return "dexec/base-c:1.0.1"
	case "c++":
		return "dexec/base-c:1.0.1"
	case "golang":
		return "golang:onbuild"
	case "python":
		return "python:3-onbuild"
	case "ruby":
		return "ruby:2.2.2-onbuild"
	case "perl":
		return "perl:5.20"
	case "php":
		return "php"
	case "clojure":
		return "clojure"
	case "haskell":
		return "haskell:latest"
	case "nodejs":
		return "node:0.12-onbuild"
	}
	return "whatever the default image is"
}

// generateRandomImageName generates a random image name that is 32 chars long.
func generateRandomImageName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
