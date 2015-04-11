package main

import (
	"encoding/json"
	"fmt"
	"github.com/acmacalister/helm"
	"github.com/libgit2/git2go"
	"github.com/maverickames/bay"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	r := helm.New(fallThrough)
	r.Handle("POST", "/web_hook", webHook)

	fmt.Println("Listing at @localhost:8080")
	http.ListenAndServe(":8080", r)
}

func fallThrough(w http.ResponseWriter, r *http.Request, params url.Values) {
	fmt.Fprintf(w, "404 - Not found")
}

func webHook(w http.ResponseWriter, r *http.Request, params url.Values) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(body))
	var wh webhook
	err = json.Unmarshal(body, &wh)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(wh)
	fmt.Fprintf(w, "ok")
	fmt.Println("Post trigger")

	bay.BuildWithGitRepo(bay.GolangEnv, wh.Repository.URL, wh.After)
	//cloneProject(wh.Repository.URL, wh.Repository.Name)
}

func cloneProject(url, name string) {
	fmt.Printf("Cloning down: %s\nUrl: %s\n", name, url)
	_, err := git.Clone(url, name, &git.CloneOptions{})
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Clone successfull")
	}
}
