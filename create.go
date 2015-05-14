package bay

import (
	"io/ioutil"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/libgit2/git2go"
)

func (s *server) buildWithGit(gitUrl, lang string) (*docker.Container, error) {
	dir, err := ioutil.TempDir("", "bay-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if _, err := git.Clone(gitUrl, dir, &git.CloneOptions{}); err != nil {
		return nil, err
	}

	return createContainer(dir, lang)

}

func (s *server) buildWithFiles(f *os.File, lang string) {

}

func (s *server) createContainer(dir, lang string) {
	volPathOpts := docker.NewPathOpts()
	volPathOpts.Set(dir)
	config := &docker.Config{
		CpuShares:       s.config.Cpu,
		Memory:          s.config.Memory,
		Tty:             true,
		OpenStdin:       false,
		Volumes:         volPathOpts,
		Cmd:             []string{dir},
		Image:           imageFromLang(lang),
		NetworkDisabled: false,
	}
	return s.dockerClient.CreateContainer(config)
}

func imageFromLang(lang string) string {
	switch lang {
	case "c":
		return "name"
	case "c++":
		return "name"
	case "golang":
		return "name"
	case "python":
		return "name"
	case "ruby":
		return "name"
	case "perl":
		return "name"
	}
	return "whatever the default image is"
}
