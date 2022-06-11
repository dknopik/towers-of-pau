package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dknopik/towersofpau"
)

func main() {
	if len(os.Args) < 2 {
		panic("invalid amount of args, need 2")
	}
	url := os.Args[1]
	client := NewClient(url)
	// Register with the coordinator
	if err := client.Register(); err != nil {
		panic(err)
	}

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
