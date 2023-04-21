/*
 */
package main

import (
	"errors"
	"fmt"

	"github.com/senzing/go-messaging/messenger"
)

var idMessages = map[int]string{
	0001: "TRACE: %s works with %s",
	1001: "DEBUG: %s works with %s",
	2001: "INFO: %s works with %s",
	3001: "WARN: %s works with %s",
	4001: "ERROR: %s works with %s",
	5001: "FATAL: %s works with %s",
	6001: "PANIC: %s works with %s",
	7001: "Xxxxx: %s works with %s",
}

func main() {
	// ctx := context.TODO()

	// Create some fake errors.

	err1 := errors.New("error #1")
	err2 := errors.New(`
	{
		"time": "2023-04-10T11:00:20.623748617-04:00",
		"level": "TRACE",
		"id": "senzing-99990002",
		"text": "A fake error",
		"location": "In main() at main.go:36",
		"details": {
			"1": "Bob",
			"2": "Mary"
		}
	}`)

	// ------------------------------------------------------------------------
	// --- Using a bare message generator
	// ------------------------------------------------------------------------

	// Example from README.md

	aMessenger, _ := messenger.New()
	fmt.Println(aMessenger.NewJson(0001, "Bob", "Mary"))
	fmt.Println(aMessenger.NewSlog(0001, "Bob", "Mary"))

	// Create a bare message generator.

	messenger1, err := messenger.New()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	// Print some messages.

	fmt.Println(messenger1.NewJson(0001, "Bob", "Mary"))
	fmt.Println(messenger1.NewJson(1001, "Bob", "Mary", err1, err2))

	// ------------------------------------------------------------------------
	// --- Using a configured message generator
	// ------------------------------------------------------------------------

	// Create a configured message generator.

	optionSenzingComponentId := &messenger.OptionSenzingComponentId{Value: 9998}
	optionCallerSkip := &messenger.OptionCallerSkip{Value: 2}
	optionIdMessages := &messenger.OptionIdMessages{Value: idMessages}

	messenger2, err := messenger.New(optionSenzingComponentId, optionCallerSkip, optionIdMessages)

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	// Print some messages.

	fmt.Println(messenger2.NewJson(0001, "Bob", "Mary"))
	fmt.Println(messenger2.NewJson(1001, "Bob", "Mary", err1, err2))

}
