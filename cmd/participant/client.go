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

type Client struct {
	url       string
	sessionID string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
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
	var info CeremonyStatus
	if err := json.Unmarshal(responseData, &info); err != nil {
		return nil, err
	}

	fmt.Printf("Ceremony status: %v\n", info)
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

func (c *Client) Contribute(contribution *towersofpau.BatchContribution) error {
	fmt.Printf("Submitting our contribution\n")
	url := fmt.Sprintf("%v/%v", c.url, "contribute")
	marshalled, err := json.Marshal(contribution)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalled))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+c.sessionID)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %v\n", string(responseData))

	if err := handleStatus(resp.StatusCode); err != nil {
		return err
	}
	return nil
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
	if err := handleStatus(resp.StatusCode); err != nil {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("Response: %v\n", string(responseData))
		return err
	}
	return nil
}

func (c *Client) TryContribute() (*towersofpau.BatchContribution, error) {
	url := fmt.Sprintf("%v/%v/%v", c.url, "lobby", "try_contribute")
	fmt.Printf("Trying to contribute at %v with ssid %v\n", url, c.sessionID)

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", "Bearer "+c.sessionID)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var errorUnm ErrorStruct
	if err := json.Unmarshal(responseData, &errorUnm); err == nil && len(errorUnm.Error) > 0 {
		return nil, fmt.Errorf("%v: %v", errorUnm.Code, errorUnm.Error)
	}

	if err := handleStatus(resp.StatusCode); err != nil {
		fmt.Printf("Response: %v", string(responseData))
		return nil, err
	}

	var contribution towersofpau.BatchContribution
	if err := json.Unmarshal(responseData, &contribution); err != nil {
		return nil, err
	}
	return &contribution, nil
}

func handleStatus(status int) error {
	switch status {
	case 200:
		return nil
	case 400:
		return errors.New("invalid ceremony")
	case 403:
		return errors.New("invalid ticket provided")
	}
	return fmt.Errorf("invalid status code: %v", status)
}
