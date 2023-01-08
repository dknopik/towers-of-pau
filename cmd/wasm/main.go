package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/savid/towersofpau"
)

type fn func(this js.Value, args []js.Value) (any, error)

var (
	jsErr     js.Value = js.Global().Get("Error")
	jsPromise js.Value = js.Global().Get("Promise")
)

func pulse() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		return true
	})
	return jsonFunc
}

func asyncFunc(innerFunc fn) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(_ js.Value, promFn []js.Value) any {
			resolve, reject := promFn[0], promFn[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(jsErr.New(fmt.Sprint("panic:", r)))
					}
				}()

				res, err := innerFunc(this, args)
				if err != nil {
					reject.Invoke(jsErr.New(err.Error()))
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		return jsPromise.New(handler)
	})
}

func parseCeramony(input string) (*towersofpau.Ceremony, error) {
	decoder := json.NewDecoder(strings.NewReader(input))
	decoder.DisallowUnknownFields()
	jsonceremony := towersofpau.JSONCeremony{
		Transcripts: []towersofpau.JSONTranscript{},
	}

	if err := decoder.Decode(&jsonceremony); err != nil {
		return nil, err
	}

	return towersofpau.DeserializeJSONCeremony(jsonceremony)
}

func contribute(this js.Value, args []js.Value) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("invalid number of arguments passed: %d", len(args))
	}

	ceramony, err := parseCeramony(args[0].String())
	if err != nil {
		return nil, err
	}

	if len(args)-1 != len(ceramony.Transcripts) {
		return nil, fmt.Errorf("invalid number of arguments passed: %d. should be %d", len(args)-1, len(ceramony.Transcripts))
	}

	for i, transcript := range ceramony.Transcripts {
		if err := towersofpau.UpdateTranscript(transcript, []byte(args[i+1].String())); err != nil {
			return nil, fmt.Errorf("failed to update transcript %d: %w", i, err)
		}
	}

	ser, err := towersofpau.SerializeJSONCeremony(ceramony)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ceremony: %w", err)
	}

	jsonString, err := json.Marshal(ser)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ceremony: %w", err)
	}

	return string(jsonString), nil
}

func main() {
	js.Global().Set("pulse", pulse())
	js.Global().Set("contribute", asyncFunc(contribute))
	select {}
}
