package bay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/acmacalister/helm"
	"github.com/fsouza/go-dockerclient"
)

// Build is type for implementing callback functions set in your config.
type Build interface {
	PreBuild(buildDir, lang string, err error)                     // PreBuild gets called before the docker image is built.
	PostBuild(container *docker.Container, lang string, err error) // PostBuild is called after the docker image is built.
}

// Config is a type for configuring the properties of bay.
type Config struct {
	CPU            int64  // Cpu is the number of CPUs allowed by each container.
	Memory         int64  // Memory is the amount of memory allowed by each container.
	DockerUrl      string // DockerUrl is the url to your docker instance. Generally your swarm url.
	Cert           string // Cert is your TLS certificate for connecting to your docker instance.
	Key            string // Key is your TLS key for connecting to your docker instance.
	Ca             string // Ca is your TLS ca for connecting to your docker instance.
	BuildInterface Build  // BuildInterface is the Build Interface.
}

// server is a internal struct used by http handlers.
type server struct {
	config       *Config
	dockerClient *docker.Client // the actual init'ed and connected docker client.
}

// response is a simple struct for writing back json messages.
type response struct {
	Response string `json:"response"`
}

// Start get this API party started. It is just the http listener to start handling routes.
func Start(address string, c *Config) error {
	client, err := docker.NewClient(c.DockerUrl) // need to add TLS options to properly support swarm
	if err != nil {
		return err
	}
	s := server{config: c, dockerClient: client}
	r := helm.New(fallThrough)
	r.POST("/github_webhook", s.githubWebhookHandler)
	r.POST("/upload", s.uploadHandler)
	r.POST("/git_url", s.gitHandler)
	r.Run(address)
	return nil // should never get here.
}

// fallThrough is an http catch all for routes that aren't handled by the API.
func fallThrough(w http.ResponseWriter, r *http.Request, params url.Values) {
	helm.RespondWithJSON(w, response{"Are you lost?"}, http.StatusNotFound)
}

// githubWebhookHandler is the http handler for github wehbook post.
func (s *server) githubWebhookHandler(w http.ResponseWriter, r *http.Request, params url.Values) {

	var wh webhook
	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&wh); err != nil {
		helm.RespondWithJSON(w, response{"Failed to decode Github payload"}, http.StatusInternalServerError)
	}

	go s.buildWithGit(wh.Repository.URL, "")

	helm.RespondWithJSON(w, response{"ok"}, http.StatusOK)
}

// uploadHandler is the http handler for file uploads.
func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	file, header, err := r.FormFile("file")
	if err != nil {
		helm.RespondWithJSON(w, response{"failed to get file from form"}, http.StatusBadRequest)
		return
	}

	lang, ok := params["language"]
	if !ok {
		helm.RespondWithJSON(w, response{"Please specific a language parameter to build with."}, http.StatusBadRequest)
		return
	}

	go s.buildWithFiles(file, lang[0], header.Header.Get("Content-Type")) // need lang from API

	helm.RespondWithJSON(w, response{fmt.Sprintf("%s uploaded successfully", header.Filename)}, http.StatusOK)
}

// gitHandler is the http handler for bare git urls.
func (s *server) gitHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	lang, ok := params["language"]
	if !ok {
		helm.RespondWithJSON(w, response{"Please specific a language parameter to build with."}, http.StatusBadRequest)
		return
	}

	gitUrl, ok := params["git_url"]
	if !ok {
		helm.RespondWithJSON(w, response{"Please specific a git_url parameter to clone with."}, http.StatusBadRequest)
		return
	}

	go s.buildWithGit(gitUrl[0], lang[0])

	helm.RespondWithJSON(w, response{"ok"}, http.StatusOK)
}
