package messenger

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var idMessages = map[int]string{
	2001: "%s knows %s",
	3001: "%s knows %s",
	4001: "%s knows %s",
	2:    "%s does not know %s",
}

var testCasesForMessage = []struct {
	name                string
	messageNumber       int
	options             []interface{}
	details             []interface{}
	expectedMessageJson string
	expectedMessageSlog []interface{}
}{
	{
		name:                "messenger-1",
		messageNumber:       1,
		options:             []interface{}{getOptionIdMessages(), getOptionCallerSkip()},
		details:             []interface{}{"A", 1, getTimestamp()},
		expectedMessageJson: `{"time":"2000-01-01 00:00:00 +0000 UTC","level":"TRACE","id":"senzing-99990001","location":"In func1() at messenger_test.go:113","details":{"1":"A","2":1}}`,
		expectedMessageSlog: []interface{}{"id", "senzing-99990001", "location", "In NewSlog() at messenger.go:367", "details", map[string]interface{}{"1": "A", "2": 1}},
	},
	{
		name:                "messenger-2",
		messageNumber:       2,
		options:             []interface{}{getOptionIdMessages(), getOptionCallerSkip()},
		details:             []interface{}{"Bob", "Jane", getTimestamp()},
		expectedMessageJson: `{"time":"2000-01-01 00:00:00 +0000 UTC","level":"TRACE","id":"senzing-99990002","text":"Bob does not know Jane","location":"In func1() at messenger_test.go:113","details":{"1":"Bob","2":"Jane"}}`,
		expectedMessageSlog: []interface{}{"id", "senzing-99990002", "location", "In NewSlog() at messenger.go:367", "details", map[string]interface{}{"1": "Bob", "2": "Jane"}},
	},
}

// ----------------------------------------------------------------------------
// Test harness
// ----------------------------------------------------------------------------

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	code := m.Run()
	err = teardown()
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(code)
}

func setup() error {
	var err error = nil
	return err
}

func teardown() error {
	var err error = nil
	return err
}

// ----------------------------------------------------------------------------
// Internal functions - names begin with lowercase letter
// ----------------------------------------------------------------------------

func testError(test *testing.T, testObject MessengerInterface, err error) {
	if err != nil {
		assert.Fail(test, err.Error())
	}
}

func getOptionIdMessages() *OptionIdMessages {
	return &OptionIdMessages{
		Value: idMessages,
	}
}

func getOptionCallerSkip() *OptionCallerSkip {
	return &OptionCallerSkip{
		Value: 2,
	}
}

func getTimestamp() *MessageTime {
	return &MessageTime{
		Value: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ----------------------------------------------------------------------------
// Test interface functions
// ----------------------------------------------------------------------------

// -- Test New() method ---------------------------------------------------------

func TestMessengerImpl_NewJson(test *testing.T) {
	for _, testCase := range testCasesForMessage {
		if len(testCase.expectedMessageJson) > 0 {
			test.Run(testCase.name+"-NewJson", func(test *testing.T) {
				testObject, err := New(testCase.options...)
				testError(test, testObject, err)
				actual := testObject.NewJson(testCase.messageNumber, testCase.details...)
				assert.Equal(test, testCase.expectedMessageJson, actual, testCase.name)
			})
		}
	}
}

func TestMessengerImpl_NewSlog(test *testing.T) {
	for _, testCase := range testCasesForMessage {
		if len(testCase.expectedMessageSlog) > 0 {
			test.Run(testCase.name+"-NewSlog", func(test *testing.T) {
				testObject, err := New(testCase.options...)
				testError(test, testObject, err)
				_, actual := testObject.NewSlog(testCase.messageNumber, testCase.details...)
				assert.Equal(test, testCase.expectedMessageSlog, actual, testCase.name)
				// assert.Equal(test, testCase.expectedMessage, actual, testCase.name)
			})
		}
	}
}

// ----------------------------------------------------------------------------
// Examples for godoc documentation
// ----------------------------------------------------------------------------

func ExampleMessengerImpl_NewJson() {
	// For more information, visit https://github.com/Senzing/go-messaging/blob/main/messenger/messenger_test.go
	example, err := New()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(example.NewJson(2001, "Bob", "Jane", getTimestamp(), getOptionCallerSkip()))
	//Output: {"time":"2000-01-01 00:00:00 +0000 UTC","level":"INFO","id":"senzing-99992001","location":"In ExampleMessengerImpl_NewJson() at messenger_test.go:144","details":{"1":"Bob","2":"Jane"}}
}
