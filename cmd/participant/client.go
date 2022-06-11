package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dknopik/towersofpau"
)

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

type Client struct {
	url          string
	registration *registration
}

type registration struct {
	Start    time.Time
	Deadline time.Time
	Ticket   string
}

func (c *Client) StartTime() *time.Time {
	if c.registration == nil {
		return nil
	}
	return &c.registration.Start
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
	var part *registration
	if err := json.Unmarshal(responseData, part); err != nil {
		return err
	}
	c.registration = part
	return nil
}

type Info struct {
	Start    time.Time
	Deadline time.Time
	Ceremony *towersofpau.Ceremony
}

func (c *Client) GetCeremony() (*Info, error) {
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
	var info *Info
	return info, json.Unmarshal(responseData, info)
}

func (c *Client) SubmitCeremony(ceremony *towersofpau.Ceremony) error {
	if c.registration == nil {
		return errors.New("no registration available")
	}
	url := fmt.Sprintf("%v/%v/%v", c.url, "participation", c.registration.Ticket)
	reader, writer := io.Pipe()
	towersofpau.Serialize(writer, ceremony)
	resp, err := http.Post(url, "text/html; charset=UTF-8", reader)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	case 400:
		return errors.New("invalid ceremony")
	case 403:
		return errors.New("invalid ticket provided")
	}
	return errors.New("invalid status code")
}
