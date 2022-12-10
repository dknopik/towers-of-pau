package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dknopik/towersofpau"
)

type Client struct {
	url          string
	registration *registration
	sessionID    string
	closeCh      chan struct{}
}

func NewClient(url string) *Client {
	return &Client{
		url:     url,
		closeCh: make(chan struct{}),
	}
}

type registration struct {
	Start    int
	Deadline int
	Ticket   string
}

func (c *Client) StartTime() *time.Time {
	if c.registration == nil {
		return nil
	}
	t := time.Unix(int64(c.registration.Start), 0)
	return &t
}

func (c *Client) Register() error {
	url := fmt.Sprintf("%v/%v", c.url, "participation")
	var body io.Reader
	resp, err := http.Post(url, "text/html; charset=UTF-8", body)
	if err != nil {
		return err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(responseData))
	var part registration
	if err := json.Unmarshal(responseData, &part); err != nil {
		return err
	}
	c.registration = &part
	fmt.Printf("Registered for ceremony at time %v\n", time.Unix(int64(part.Start), 0))
	return nil
}

type Info struct {
	Start    int
	Deadline int
	Ceremony *towersofpau.JSONCeremony
}

func (c *Client) GetCeremony() (*Info, error) {
	fmt.Println("Fetching Ceremony")
	if c.registration == nil {
		return nil, errors.New("no registration available")
	}
	url := fmt.Sprintf("%v/%v/%v", c.url, "participation", c.registration.Ticket)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(responseData, &info); err != nil {
		return nil, err
	}

	c.registration.Start = info.Start
	c.registration.Deadline = info.Deadline

	fmt.Println("Retrieved ceremony")
	return &info, nil
}

func (c *Client) SubmitCeremony(ceremony *towersofpau.Ceremony) error {
	fmt.Println("Submitting ceremony")
	if c.registration == nil {
		return errors.New("no registration available")
	}
	url := fmt.Sprintf("%v/%v/%v", c.url, "participation", c.registration.Ticket)

	buf := new(bytes.Buffer)
	towersofpau.Serialize(buf, ceremony)
	resp, err := http.Post(url, "text/html", buf)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		fmt.Println("Submitted ceremony successfully")
		return nil
	case 400:
		return errors.New("invalid ceremony")
	case 403:
		return errors.New("invalid ticket provided")
	}
	return errors.New("invalid status code")
}
