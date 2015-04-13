package bay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/acmacalister/helm"
)

// server contains all the info for running the API endpoints
// and web dashboard.
type server struct {
	workers []string //addresses to each of the workers.
}

// Start takes a slice of worker's addresses (e.g http://worker1:8080)
// and a listener for which port to run this server now.
func Start(workers []string, listener string) {
	s := new(server)
	s.workers = append(s.workers, workers...)
	r := helm.New(fallThrough)
	r.POST("/web_hook", s.webHookHandler)
	r.Run(listener)
}

// fallThrough is the caught all route for endpoints called without a handler.
func fallThrough(w http.ResponseWriter, r *http.Request, params url.Values) {
	fmt.Fprintf(w, "404 - Not found")
}

// webHookHandler takes the response from Github webhook callbacks.
func (s *server) webHookHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	var wh webhook
	err = json.Unmarshal(body, &wh)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "ok")
	fmt.Println("Post trigger")
	// we will need some logic to pick which worker will get the job.
	// for now it is just the first worker.
	http.Post(s.workers[0], bodyType, body)
}
