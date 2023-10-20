package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dknopik/towersofpau"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("invalid number of arguments")
		return
	}
	if err := verify(os.Args[1]); err == nil {
		fmt.Println("Ceremony verified correctly")
	} else {
		fmt.Printf("Encountered error while verifying: %v\n", err)
	}
}

func verify(path string) error {
	fmt.Printf("Verifying ceremony at %v\n", path)
	var jsonCeremony towersofpau.JSONCeremony
	fmt.Printf("Reading file\n")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Printf("Unmarshalling\n")
	if err := json.Unmarshal(data, &jsonCeremony); err != nil {
		return err
	}
	fmt.Printf("Deserializing\n")
	ceremony, err := towersofpau.DeserializeJSONCeremony(jsonCeremony)
	if err != nil {
		return err
	}
	fmt.Printf("Verifying\n")
	return towersofpau.VerifyCeremony(ceremony)
}
