package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dknopik/towersofpau"
)

func (c *Client) GetStatus() (*CeremonyStatus, error) {
	url := fmt.Sprintf("%v/%v/%v", c.url, "info", "status")
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info CeremonyStatus
	if err := json.Unmarshal(responseData, &info); err != nil {
		return nil, err
	}

	fmt.Println("Ceremony status received")
	return &info, nil
}

func (c *Client) GetCurrentState() (*towersofpau.Ceremony, error) {
	url := fmt.Sprintf("%v/%v/%v", c.url, "info", "current_state")
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info towersofpau.Ceremony
	if err := json.Unmarshal(responseData, &info); err != nil {
		return nil, err
	}

	fmt.Println("Ceremony status received")
	return &info, nil
}

func (c *Client) Contribute(ceremony *towersofpau.Ceremony) error {
	url := fmt.Sprintf("%v/%v", c.url, "contribute")

	buf := new(bytes.Buffer)
	towersofpau.Serialize(buf, ceremony)
	resp, err := http.Post(url, "text/html", buf)
	if err != nil {
		return err
	}
	return handleStatus(resp.StatusCode)
}

func (c *Client) Abort(p *Participant) error {
	url := fmt.Sprintf("%v/%v/%v", c.url, "contribution", "abort")

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(p); err != nil {
		return err
	}
	resp, err := http.Post(url, "text/html", buf)
	if err != nil {
		return err
	}
	return handleStatus(resp.StatusCode)
}

func (c *Client) TryContribute(status *ParticipantQueueStatus) error {
	url := fmt.Sprintf("%v/%v/%v", c.url, "lobby", "try_contribute")

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(status); err != nil {
		return err
	}
	resp, err := http.Post(url, "text/html", buf)
	if err != nil {
		return err
	}
	return handleStatus(resp.StatusCode)
}

func (c *Client) RequestLink() error {
	return nil
}

func handleStatus(status int) error {
	switch status {
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
