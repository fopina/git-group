package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GitlabClient implements methods to interace with a Gitlab instance
type GitlabClient struct {
	Client   *http.Client
	Endpoint *url.URL
	Group    string
	Token    string
}

type gitlabLoginResponse struct {
	AccessToken      string `json:"access_token"`
	ErrorDescription string `json:"error_description"`
}

type gitlabProject struct {
	ID     int
	Name   string
	SSHURL string `json:"ssh_url_to_repo"`
	WEBURL string `json:"web_url"`
}

type optionalInt struct {
	Int  int
	Null bool
}

// ListedProject wraps a gitlabProject with iteration status to be sent through the ListGroupProjects channel
type ListedProject struct {
	Index   int
	Total   optionalInt
	Project gitlabProject
}

// NewGitlabClient returns a GitlabClient
func NewGitlabClient(gitlabLink string) (*GitlabClient, error) {
	projectURL, err := url.Parse(gitlabLink)
	if err != nil {
		return nil, err
	}

	baseURL := &url.URL{
		Scheme: projectURL.Scheme,
		Host:   projectURL.Host,
	}

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	return &GitlabClient{
		Client:   netClient,
		Endpoint: baseURL,
		Group:    strings.Trim(projectURL.Path, "/"),
	}, nil
}

// Authenticate authenticates against a Gitlab instance, storing the access_token
func (c *GitlabClient) Authenticate(username, password string) error {
	c.Endpoint.Path = "oauth/token"

	data := url.Values{
		"grant_type": {"password"},
		"username":   {username},
		"password":   {password},
	}

	resp, err := c.Client.PostForm(c.Endpoint.String(), data)

	if err != nil {
		return err
	}

	var res gitlabLoginResponse

	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		return err
	}

	if res.AccessToken != "" {
		c.Token = res.AccessToken
		return nil
	}

	if res.ErrorDescription != "" {
		return fmt.Errorf("login failed: %v", res.ErrorDescription)
	}

	return fmt.Errorf("no token nor error returned....?")
}

// ListGroupProjects lists all projects within a group
func (c *GitlabClient) ListGroupProjects(projects chan interface{}) error {
	return c.ListGroupProjectsWithMax(projects, 0)
}

// ListGroupProjectsWithMax lists all projects within a group
func (c *GitlabClient) ListGroupProjectsWithMax(projects chan interface{}, sample int) error {
	c.Endpoint.Path = "api/v4/groups/" + strings.ReplaceAll(c.Group, "/", "%2F") + "/projects"
	// FIXME: url.URL.String() uses EscapedPath() (instead of RawPath()) and re-escapes %2F...
	endpoint := strings.ReplaceAll(c.Endpoint.String(), "%252F", "%2F")

	page := "1"
	index := 1
	perPage := "100"
	if sample > 0 && sample < 100 {
		perPage = strconv.Itoa(sample)
	}

	for page != "" {
		data := url.Values{
			"page":     {page},
			"per_page": {perPage},
			"simple":   {"1"},
			"archived": {"0"},
		}

		req, err := http.NewRequest("GET", endpoint+"?"+data.Encode(), nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("gitlab returned %v", resp.StatusCode)
		}

		total := optionalInt{0, true}
		if len(resp.Header["X-Total"]) > 0 {
			total.Int, err = strconv.Atoi(resp.Header["X-Total"][0])
			total.Null = err != nil
		}

		if len(resp.Header["X-Next-Page"]) > 0 {
			page = resp.Header["X-Next-Page"][0]
		} else {
			page = ""
		}

		var res []gitlabProject

		err = json.NewDecoder(resp.Body).Decode(&res)

		if err != nil {
			return err
		}

		for _, project := range res {
			projects <- ListedProject{Index: index, Total: total, Project: project}
			index++
			sample--
			// if sample is (starts with) 0, there is no limit
			if sample == 0 {
				return nil
			}
		}
	}

	return nil
}
