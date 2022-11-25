package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
)

func authenticate() {
	/*
		siwe.InitMessage()

		message, err := siwe.ParseMessage(messageStr)
		message.

		publicKey, err := message.VerifyEIP191(signature)

		ok, err := message.ValidNow()

		publicKey, err := message.Verify(signature, optionalNonce, optionalTimestamp)
	*/
}

func (c *Client) Login() error {
	links, err := c.RequestAuthLink()
	if err != nil {
		return err
	}

	if err := startSIWEServer(c); err != nil {
		return err
	}

	if err := startSIWEPage(links.EthAuthURL); err != nil {
		return err
	}
	<-c.closeCh
	return nil
}

type AuthLinks struct {
	EthAuthURL    string `json:"eth_auth_url"`
	GithubAuthURL string `json:"github_auth_url"`
}

func (c *Client) RequestAuthLink() (*AuthLinks, error) {
	fmt.Println("Requesting Authentication")
	url := fmt.Sprintf("%v/%v/%v?redirect_to=%v", c.url, "auth", "request_link", "http://127.0.0.1:3000/auth/callback/eth")

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

func (c *Client) handleCallback(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("Received callback")
	sessionid := req.URL.Query().Get("session_id")
	if sessionid == "" {
		panic(fmt.Errorf("invalid query params %v", req.URL.Query().Encode()))
	}
	c.sessionCh <- sessionid

}

func startSIWEServer(client *Client) error {
	fmt.Println("Starting SIWE server on port 3000")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/auth/callback/eth", client.handleCallback).
		Methods("GET")
	go func() {
		if err := http.ListenAndServe(":3000", router); err != nil {
			panic(err)
		}
	}()
	return nil
}

func startSIWEPage(url string) error {
	fmt.Printf("Signing in with ethereum on %v\n", url)
	cmd := exec.Command("chromium", url)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
