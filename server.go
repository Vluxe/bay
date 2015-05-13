package bay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/acmacalister/helm"
	"github.com/gorilla/websocket"
)

const (
	bufferSize = 1024
)

// server contains all the info for running the API endpoints
// and web dashboard.
type server struct {
	workers []clientWorker //addresses to each of the workers.
}

type clientWorker struct {
	url         *url.URL    //store this for reconnect
	headers     http.Header //store this for reconnect
	conn        *websocket.Conn
	isConnected bool
}

// Start takes a slice of worker's addresses (e.g http://worker1:8080)
// and a listener for which port to run this server now.
func Start(workersAddresses []string, listener string) error {
	s := new(server)
	s.connectToWorkers(workersAddresses)
	r := helm.New(fallThrough)
	r.POST("/web_hook", s.webHookHandler)
	r.Run(listener)
	return nil
}

func (s *server) connectToWorkers(workerAddresses []string) error {
	for _, worker := range workerAddresses {
		u, err := url.Parse(worker)
		if err != nil {
			return errors.New("couldn't connect to " + worker)
		}

		websocketProtocol := "chat, superchat"
		header := make(http.Header)
		header.Add("Sec-WebSocket-Protocol", websocketProtocol)
		header.Add("Origin", u.String())

		conn, err := net.Dial("tcp", u.Host)
		if err != nil {
			return errors.New("couldn't reach " + worker)
		}

		webConn, _, err := websocket.NewClient(conn, u, header, bufferSize, bufferSize)
		if err != nil {
			return errors.New("couldn't create websocket to " + worker)
		}

		w := clientWorker{url: u, headers: header, conn: webConn, isConnected: true}
		s.workers = append(s.workers, w)
	}

	return nil
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
	//	http.Post(s.workers[0], bodyType, body)
}
