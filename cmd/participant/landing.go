package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dknopik/towersofpau"
)

type ErrorStruct struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

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
	fmt.Println(string(responseData))
	var info CeremonyStatus
	if err := json.Unmarshal(responseData, &info); err != nil {
		return nil, err
	}

	fmt.Println("Ceremony status received")
	return &info, nil
}

func (c *Client) GetCurrentState() (*towersofpau.Ceremony, error) {
	url := fmt.Sprintf("%v/%v/%v", c.url, "info", "current_state")
	fmt.Printf("Getting current state from %v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	defer func() {
		fmt.Printf("Deserializing took %v\n", time.Since(start))
	}()
	return towersofpau.Deserialize(bytes.NewReader(responseData))
}

func (c *Client) Contribute(ceremony *towersofpau.Ceremony) error {
	url := fmt.Sprintf("%v/%v", c.url, "contribute")
	buf := new(bytes.Buffer)
	towersofpau.Serialize(buf, ceremony)

	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.sessionID)
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
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

func (c *Client) TryContribute() error {
	url := fmt.Sprintf("%v/%v/%v", c.url, "lobby", "try_contribute")

	fmt.Printf("Trying to contribute at %v with ssid %v\n", url, c.sessionID)

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.sessionID)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var errorUnm ErrorStruct
	if err := json.Unmarshal(responseData, &errorUnm); err == nil {
		return fmt.Errorf("%v: %v", errorUnm.Code, errorUnm.Error)
	}

	fmt.Println(string(responseData))
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
	return fmt.Errorf("invalid status code: %v", status)
}
