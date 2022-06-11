package main

import (
	"errors"
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
	// Retrieve our start time
	start := client.StartTime()
	if start == nil {
		panic("invalid start time")
	}
	// Wait for our start time
	time.Sleep(time.Until(*start))
	// Get the ceremony
	info, err := client.GetCeremony()
	if err != nil {
		panic(err)
	}
	// Participate
	newCeremony := info.Ceremony.Copy()
	if err := participate(newCeremony); err != nil {
		panic(err)
	}
	// Send reply
	if err := client.SubmitCeremony(newCeremony); err != nil {
		panic(err)
	}
}

func participate(ceremony *towersofpau.Ceremony) error {
	// Verify the data
	if !towersofpau.SubgroupChecksParticipant(ceremony) {
		return errors.New("subgroup check failed")
	}
	// Add our contribution
	return towersofpau.UpdateTranscript(ceremony)
}
