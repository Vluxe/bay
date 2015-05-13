package bay

import (
	"time"
)

type webhook struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Created    bool       `json:"created"`
	Deleted    bool       `json:"deleted"`
	Forced     bool       `json:"forced"`
	Commits    []commit   `json:"commits"`
	Repository repository `json:"repository"`
}

type commit struct {
	Id        string    `json:"id"`
	Distinct  bool      `json:"distinct"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Auther    User      `json:"author"`
	Committer User      `json:"committer"`
}

type repository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    User   `json:"owner"`
	HtmlURL  string `json:"html_url"`
	URL      string `json:"url"`
}

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
