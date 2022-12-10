package main

import (
	"errors"
	"fmt"
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
	go func() {
		if err := client.Login(); err != nil {
			panic(err)
		}
	}()

	if err := runParticipation(client); err != nil {
		panic(err)
	}
}

func runParticipation(client *Client) error {
	// Wait for our sessionID
	client.WaitForSessionID()

	// Register with the coordinator
	_, err := client.GetStatus()
	if err != nil {
		panic(err)
	}

	attemptCount := 0
	for {
		err := client.TryContribute()
		if err == nil {
			break
		}
		attemptCount += 1
		if strings.HasPrefix(err.Error(), "TryContributeError::AnotherContributionInProgress") {
			fmt.Fprintf(os.Stderr, "Another contribution is in progress, trying again in 15 seconds (attempt %v)\n", attemptCount)
		} else {
			fmt.Fprintf(os.Stderr, "Error attempting to contribute: %v (attempt %v)\n", err, attemptCount)
		}
		time.Sleep(15 * time.Second)
	}

	ceremony, err := client.GetCurrentState()
	if err != nil {
		return err
	}

	newCeremony := ceremony.Copy()

	if err := participate(newCeremony); err != nil {
		return err
	}

	if err := towersofpau.VerifySubmission(ceremony, newCeremony); err != nil {
		return err
	}

	if err := client.Contribute(newCeremony); err != nil {
		return err
	}

	/*
		var info *Info
		for info == nil || info.Ceremony == nil {
			// Retrieve our start time
			start := client.StartTime()
			if start == nil {
				panic("invalid start time")
			}
			// Wait for our start time
			fmt.Printf("Waiting for our start time: %v\n", time.Until(*start))
			time.Sleep(time.Until(*start))
			// Get the ceremony
			var err error
			info, err = client.GetCeremony()
			if err != nil {
				panic(err)
			}
		}

		ceremony, err := towersofpau.DeserializeJSONCeremony(*info.Ceremony)
		if err != nil {
			panic(err)
		}

		// Participate
		newCeremony := ceremony.Copy()
		if err := participate(newCeremony); err != nil {
			panic(err)
		}
		// Send reply
		if err := client.SubmitCeremony(newCeremony); err != nil {
			panic(err)
		}
	*/
	close(client.closeCh)
	return nil
}

func participate(ceremony *towersofpau.Ceremony) error {
	fmt.Println("Calculating our contribution")
	start := time.Now()
	// Verify the data
	if !towersofpau.SubgroupChecksParticipant(ceremony) {
		return errors.New("subgroup check failed")
	}
	// Add our contribution
	if err := towersofpau.UpdateTranscript(ceremony); err != nil {
		return err
	}
	fmt.Printf("Contribution calculated in %v\n", time.Since(start))
	return nil
}
