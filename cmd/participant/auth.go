package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func (c *Client) Login() error {
	links, err := c.RequestAuthLink()
	if err != nil {
		return err
	}

	if err := startSIWEPage(links.EthAuthURL); err != nil {
		return err
	}

	fmt.Println("Please paste the session cookie from the browser")
	token, err := readCookie()
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated as %v with ssid %v\n", token.IDtoken.Nickname, token.SessionID)
	c.sessionID = token.SessionID
	return nil
}

type AuthLinks struct {
	EthAuthURL    string `json:"eth_auth_url"`
	GithubAuthURL string `json:"github_auth_url"`
}

func (c *Client) RequestAuthLink() (*AuthLinks, error) {
	fmt.Println("Requesting Authentication")
	url := fmt.Sprintf("%v/%v/%v", c.url, "auth", "request_link")

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var links AuthLinks
	if err := json.Unmarshal(responseData, &links); err != nil {
		return nil, err
	}
	return &links, nil
}

func startSIWEPage(url string) error {
	fmt.Printf("Signing in with ethereum on %v\n", url)
	cmd := exec.Command("chromium", url)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

type cookie struct {
	Expiry   uint64 `json:"exp"`
	Nickname string `json:"nickname"`
	Provider string `json:"provider"`
}

type token struct {
	IDtoken   cookie `json:"id_token"`
	SessionID string `json:"session_id"`
}

func readCookie() (*token, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var t token
	if err := json.Unmarshal(input, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
