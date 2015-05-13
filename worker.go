package bay

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/acmacalister/helm"
)

// worker contains all the info needed for starting a worker.
type worker struct {
	address string
}

// StartWorker starts a worker. Currently it starts an HTTP server
// to "talk" with the frontend server, but a WebSocket might be a better call.
func StartWorker(address string) {
	//w := worker{address: address}
	r := helm.New(fallThrough)
	r.POST("/gitJob", gitJobHandler)
	r.Run(address)
}

func gitJobHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	//BuildWithGitRepo(gitUrl, commitId)
	fmt.Fprintf(w, "ok")
}
