package messenger

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
)

// ----------------------------------------------------------------------------
// Types - interface
// ----------------------------------------------------------------------------

// The MessengerInterface interface is...
type MessengerInterface interface {
	NewJson(messageNumber int, details ...interface{}) string
	NewSlog(messageNumber int, details ...interface{}) (string, []interface{})
	NewSlogLevel(messageNumber int, details ...interface{}) (string, slog.Level, []interface{})
}

// ----------------------------------------------------------------------------
// Types - struct
// ----------------------------------------------------------------------------

// Fields in the formatted message.
// Order is important.
// It should be time, level, id, text, status, duration, location, errors, details.
type MessageFormat struct {
	Time     string      `json:"time,omitempty"`     // Time of message in UTC.
	Level    string      `json:"level,omitempty"`    // Level:  TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC.
	Id       string      `json:"id,omitempty"`       // Message identifier.
	Text     interface{} `json:"text,omitempty"`     // Message text.
	Status   string      `json:"status,omitempty"`   // Status information.
	Duration int64       `json:"duration,omitempty"` // Duration in nanoseconds
	Location string      `json:"location,omitempty"` // Location in the code issuing message.
	Errors   interface{} `json:"errors,omitempty"`   // List of errors.
	Details  interface{} `json:"details,omitempty"`  // All instances passed into the message.
}

// --- Override values when creating messages ---------------------------------

type MessageDetails struct {
	Value interface{}
}

type MessageDuration struct {
	Value int64
}

type MessageId struct {
	Value string
}

type MessageLevel struct {
	Value string
}

type MessageLocation struct {
	Value string
}

type MessageStatus struct {
	Value string
}

type MessageText struct {
	Value interface{}
}

type MessageTime struct {
	Value time.Time
}

// --- Options for New() ------------------------------------------------------

type OptionCallerSkip struct {
	Value int
}

type OptionIdMessages struct {
	Value map[int]string
}

type OptionIdStatuses struct {
	Value map[int]string
}

type OptionSenzingProductId struct {
	Value int
}

type OptionMessageIdTemplate struct {
	Value string
}

// ----------------------------------------------------------------------------
// Constants
// ----------------------------------------------------------------------------

// Log levels as integers.
// Compatible with golang.org/x/exp/slog.
const (
	LevelTraceInt int = -8
	LevelDebugInt int = -4
	LevelInfoInt  int = 0
	LevelWarnInt  int = 4
	LevelErrorInt int = 8
	LevelFatalInt int = 12
	LevelPanicInt int = 16
)

// Strings representing the supported logging levels.
const (
	LevelDebugName = "DEBUG"
	LevelErrorName = "ERROR"
	LevelFatalName = "FATAL"
	LevelInfoName  = "INFO"
	LevelPanicName = "PANIC"
	LevelTraceName = "TRACE"
	LevelWarnName  = "WARN"
)

// Existing and new log levels used with slog.Level.
const (
	LevelDebugSlog = slog.LevelDebug
	LevelErrorSlog = slog.LevelError
	LevelFatalSlog = slog.Level(LevelFatalInt)
	LevelInfoSlog  = slog.LevelInfo
	LevelPanicSlog = slog.Level(LevelPanicInt)
	LevelTraceSlog = slog.Level(LevelTraceInt)
	LevelWarnSlog  = slog.LevelWarn
)

// ----------------------------------------------------------------------------
// Variables
// ----------------------------------------------------------------------------

// Message ID Low-bound for message levels
// i.e. a message in range 0 - 999 is a TRACE message.
var IdLevelRangesAsString = map[int]string{
	0000: LevelTraceName,
	1000: LevelDebugName,
	2000: LevelInfoName,
	3000: LevelWarnName,
	4000: LevelErrorName,
	5000: LevelFatalName,
	6000: LevelPanicName,
}

// Map from slog.Level to string representation.
var LevelToTextMap = map[slog.Level]string{
	LevelDebugSlog: LevelDebugName,
	LevelErrorSlog: LevelErrorName,
	LevelFatalSlog: LevelFatalName,
	LevelInfoSlog:  LevelInfoName,
	LevelPanicSlog: LevelPanicName,
	LevelTraceSlog: LevelTraceName,
	LevelWarnSlog:  LevelWarnName,
}

// Map from string representation to Log level as typed integer.
var TextToLevelMap = map[string]slog.Level{
	LevelDebugName: LevelDebugSlog,
	LevelErrorName: LevelErrorSlog,
	LevelFatalName: LevelFatalSlog,
	LevelInfoName:  LevelInfoSlog,
	LevelPanicName: LevelPanicSlog,
	LevelTraceName: LevelTraceSlog,
	LevelWarnName:  LevelWarnSlog,
}

// ----------------------------------------------------------------------------
// Public functions
// ----------------------------------------------------------------------------

/*
The New function creates a new instance of MessengerInterface.
Adding options can be used to modify subcomponents.
*/
func New(options ...interface{}) (MessengerInterface, error) {
	var err error = nil
	var result MessengerInterface = nil

	// Default values.

	var (
		callerSkip        int            = 0
		idMessages        map[int]string = map[int]string{}
		idStatuses        map[int]string = map[int]string{}
		productIdentifier int            = 9999
		messageIdTemplate string         = fmt.Sprintf("senzing-%04d", productIdentifier) + "%04d"
	)

	// Process options.

	for _, value := range options {
		switch typedValue := value.(type) {
		case *OptionCallerSkip:
			callerSkip = typedValue.Value
		case *OptionIdMessages:
			idMessages = typedValue.Value
		case *OptionIdStatuses:
			idStatuses = typedValue.Value
		case *OptionSenzingProductId:
			productIdentifier = typedValue.Value
			messageIdTemplate = fmt.Sprintf("senzing-%04d", productIdentifier) + "%04d"
		case *OptionMessageIdTemplate:
			messageIdTemplate = typedValue.Value + "%04d"
		}
	}

	// Detect incorrect option values.

	if productIdentifier <= 0 || productIdentifier >= 10000 {
		err := errors.New("productIdentifier must be in range 1..9999. See https://github.com/Senzing/knowledge-base/blob/main/lists/senzing-product-ids.md")
		return result, err
	}

	if idMessages == nil {
		err := errors.New("messages must be a map[int]string")
		return result, err
	}

	if idStatuses == nil {
		err := errors.New("statuses must be a map[int]string")
		return result, err
	}

	// Create MessengerInterface.

	result = &MessengerImpl{
		callerSkip:        callerSkip,
		idMessages:        idMessages,
		idStatuses:        idStatuses,
		messageIdTemplate: messageIdTemplate,
	}
	return result, err
}

/*
The HandlerOptions function returns a slog handler that includes TRACE, FATAL, and PANIC.
TODO: Move to Senzing/go-logging.
See: https://go.googlesource.com/exp/+/refs/heads/master/slog/example_custom_levels_test.go
*/
func HandlerOptions(leveler slog.Leveler) *slog.HandlerOptions {
	if leveler == nil {
		leveler = LevelInfoSlog
	}
	handlerOptions := &slog.HandlerOptions{
		Level: leveler,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "text"
			}
			if a.Key == slog.LevelKey {
				level := ""
				switch typedValue := a.Value.Any().(type) {
				case string:
					level = typedValue
				case slog.Level:
					level = typedValue.String()
				}
				switch {
				case level == "DEBUG-4":
					a.Value = slog.StringValue("TRACE")
				case level == "ERROR+4":
					a.Value = slog.StringValue("FATAL")
				case level == "ERROR+8":
					a.Value = slog.StringValue("PANIC")
				}
			}
			return a
		},
	}
	return handlerOptions
}
