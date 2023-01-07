package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/savid/towersofpau"
)

func parseCeramony(input string) (*towersofpau.Ceremony, error) {
	var jsonceremony towersofpau.JSONCeremony
	decoder := json.NewDecoder(strings.NewReader(input))
	if err := decoder.Decode(&jsonceremony); err != nil {
		return nil, err
	}
	return towersofpau.DeserializeJSONCeremony(jsonceremony)
}

func initCeramony() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		_, err := parseCeramony(inputJSON)
		if err != nil {
			fmt.Printf("unable to parse ceramony json %s\n", err)
			return err.Error()
		}
		return true
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("init", initCeramony())
	<-make(chan bool)
}
