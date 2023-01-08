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

func generatePotPubkey(this js.Value, args []js.Value) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments passed: %d", len(args))
	}

	secret := []byte(args[0].String())
	if len(secret) < 1 {
		return nil, fmt.Errorf("invalid secret")
	}

	potPubkey := towersofpau.NewPotPubkey(secret)

	return towersofpau.SerializeJSONPointG2(potPubkey), nil
}

func contribute(this js.Value, args []js.Value) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("invalid number of arguments passed: %d", len(args))
	}

	contribution, err := towersofpau.Deserialize(strings.NewReader(args[0].String()))
	if err != nil {
		return nil, err
	}

	secret := []byte(args[1].String())
	if len(secret) < 1 {
		return nil, fmt.Errorf("invalid secret")
	}

	if err := towersofpau.UpdateContribution(contribution, secret); err != nil {
		return nil, fmt.Errorf("failed to update contribution: %w", err)
	}

	ser, err := towersofpau.SerializeJSONContribution(contribution)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize contribution: %w", err)
	}

	jsonString, err := json.Marshal(ser)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contribution: %w", err)
	}

	return string(jsonString), nil
}

func main() {
	js.Global().Set("generatePotPubkey", asyncFunc(generatePotPubkey))
	js.Global().Set("contribute", asyncFunc(contribute))
	select {}
}
