package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dknopik/towersofpau"
)

func main() {
	if len(os.Args) < 2 {
		panic("invalid amount of args, need 2")
	}
	url := os.Args[1]
	client := NewClient(url)

	// start the login process
	if err := client.Login(); err != nil {
		panic(err)
	}

	if err := runParticipation(client); err != nil {
		panic(err)
	}
}

func runParticipation(client *Client) error {
	// Register with the coordinator
	_, err := client.GetStatus()
	if err != nil {
		panic(err)
	}
	var (
		attemptCount = 0
		contribution *towersofpau.BatchContribution
	)
	/*
		marshalled, err := os.ReadFile("temp.json")
		if err != nil {
			return err
		}
		var c towersofpau.BatchContribution
		if err := json.Unmarshal(marshalled, &c); err != nil {
			return err
		}
		marshalled2, err := json.Marshal(contribution)
		if err != nil {
			return err
		}
		os.WriteFile("temp2.json", marshalled2, fs.ModeAppend)
	*/
	for {
		contribution, err = client.TryContribute()
		if err == nil {
			break
		}
		attemptCount += 1
		if strings.HasPrefix(err.Error(), "TryContributeError::AnotherContributionInProgress") {
			fmt.Fprintf(os.Stderr, "Another contribution is in progress, trying again in 15 seconds (attempt %v)\n", attemptCount)
		} else {
			fmt.Fprintf(os.Stderr, "Error attempting to contribute: %v (attempt %v)\n", err, attemptCount)
		}
		time.Sleep(30 * time.Second)
	}

	newCeremony := contribution.Copy()
	if err := participate(newCeremony); err != nil {
		return err
	}

	if err := client.Contribute(newCeremony); err != nil {
		return err
	}

	marshalled, err := json.Marshal(newCeremony)
	if err != nil {
		return err
	}
	ioutil.WriteFile("contribution.json", marshalled, fs.ModeAppend | 0o644)

	fmt.Println("Successfully contributed, exiting")
	return nil
}

func participate(ceremony *towersofpau.BatchContribution) error {
	fmt.Println("Calculating our contribution")
	start := time.Now()
	// Verify the data
	if !ceremony.SubgroupChecks() {
		return errors.New("subgroup check failed")
	}
	// Add our contribution
	if err := towersofpau.UpdateContribution(ceremony); err != nil {
		return err
	}
	fmt.Printf("Contribution calculated in %v\n", time.Since(start))
	return nil
}
