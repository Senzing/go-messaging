package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

// MessengerImpl is an example type-struct.
type MessengerImpl struct {
	idMessages          map[int]string // Map message numbers to text format strings
	idStatuses          map[int]string
	messageIdTemplate   string // A string template for fmt.Sprinf()
	callerSkip          int    // Levels of code nexting to skip when calculation location
	sortedIdLevelRanges []int  // The keys of IdLevelRanges in sorted order.
}

// ----------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------

// Determine if string is syntactically JSON.
func isJson(unknownString string) bool {
	unknownStringUnescaped, err := strconv.Unquote(unknownString)
	if err != nil {
		unknownStringUnescaped = unknownString
	}
	var jsonString json.RawMessage
	return json.Unmarshal([]byte(unknownStringUnescaped), &jsonString) == nil
}

// Make JSON string into an interface{}.
func jsonAsInterface(unknownString string) interface{} {
	unknownStringUnescaped, err := strconv.Unquote(unknownString)
	if err != nil {
		unknownStringUnescaped = unknownString
	}
	var jsonString json.RawMessage
	json.Unmarshal([]byte(unknownStringUnescaped), &jsonString)
	return jsonString
}

// Make an interface{} into a string.
func interfaceAsString(unknown interface{}) string {
	// See https://pkg.go.dev/fmt for format strings.
	var result string

	switch value := unknown.(type) {
	case nil:
		result = "<nil>"
	case string:
		result = value
	case int:
		result = fmt.Sprintf("%d", value)
	case float64:
		result = fmt.Sprintf("%g", value)
	case bool:
		result = fmt.Sprintf("%t", value)
	case error:
		result = value.Error()
	default:
		// xType := reflect.TypeOf(unknown)
		// xValue := reflect.ValueOf(unknown)
		// result = fmt.Sprintf("(%s)%#v", xType, xValue)
		result = fmt.Sprintf("%#v", unknown)
	}
	return result
}

// Walk through the details to improve their future JSON representation.
func messageDetails(details ...interface{}) interface{} {

	result := make(map[string]interface{})

	// Process different types of details.

	for index, value := range details {
		switch typedValue := value.(type) {
		case nil:
			result[strconv.Itoa(index+1)] = "<nil>"

		case int, float64:
			result[strconv.Itoa(index+1)] = typedValue

		case string:
			if isJson(typedValue) {
				result[strconv.Itoa(index+1)] = jsonAsInterface(typedValue)
			} else {
				result[strconv.Itoa(index+1)] = typedValue
			}

		case bool:
			result[strconv.Itoa(index+1)] = fmt.Sprintf("%t", typedValue)

		case error:
			// do nothing.

		case map[string]string:
			for mapIndex, mapValue := range typedValue {
				mapValueAsString := interfaceAsString(mapValue)
				if isJson(mapValueAsString) {
					result[mapIndex] = jsonAsInterface(mapValueAsString)
				} else {
					result[mapIndex] = mapValueAsString
				}
			}

		default:
			valueAsString := interfaceAsString(typedValue)
			if isJson(valueAsString) {
				result[strconv.Itoa(index+1)] = jsonAsInterface(valueAsString)
			} else {
				result[strconv.Itoa(index+1)] = valueAsString
			}
		}
	}

	if len(result) == 0 {
		result = nil
	}

	return result
}

// ----------------------------------------------------------------------------
// Private methods
// ----------------------------------------------------------------------------

// Create a slice of ["key1", value1, "key2", value2, ...] which has oscillating
// key and values in the slice.
func (messenger *MessengerImpl) getKeyValuePairs(appMessageFormat *MessageFormat, keys []string) []interface{} {
	var result []interface{} = nil
	keyValueMap := map[string]interface{}{
		"time":     appMessageFormat.Time,
		"level":    appMessageFormat.Level,
		"id":       appMessageFormat.Id,
		"status":   appMessageFormat.Status,
		"duration": appMessageFormat.Duration,
		"location": appMessageFormat.Location,
		"errors":   appMessageFormat.Errors,
		"details":  appMessageFormat.Details,
	}

	// In key order, append values to result.

	for _, key := range keys {
		value, ok := keyValueMap[key]
		if !ok {
			continue
		}
		switch typedValue := value.(type) {
		case string:
			if typedValue != "" {
				result = append(result, key, value)
			}
		case int64:
			if typedValue != 0 {
				result = append(result, key, value)
			}
		default:
			if typedValue != nil {
				result = append(result, key, value)
			}
		}
	}
	return result
}

// Given a message number, figure out the Level (TRACE, DEBUG, ..., FATAL, PANIC)
func (messenger *MessengerImpl) getLevel(messageNumber int) string {
	sortedMessageLevelKeys := messenger.getSortedIdLevelRanges(IdLevelRangesAsString)
	for _, messageLevelKey := range sortedMessageLevelKeys {
		if messageNumber >= messageLevelKey {
			return IdLevelRangesAsString[messageLevelKey]
		}
	}
	return "UNKNOWN"
}

// Since a map[int]any is not guaranteed to be in order, return an ordered slice of int.
func (messenger *MessengerImpl) getSortedIdLevelRanges(idLevelRanges map[int]string) []int {
	if messenger.sortedIdLevelRanges == nil {
		messenger.sortedIdLevelRanges = make([]int, 0, len(idLevelRanges))
		for key := range idLevelRanges {
			messenger.sortedIdLevelRanges = append(messenger.sortedIdLevelRanges, key)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(messenger.sortedIdLevelRanges)))
	}
	return messenger.sortedIdLevelRanges
}

// Create a populated MessageFormat.
func (messenger *MessengerImpl) populateStructure(messageNumber int, details ...interface{}) *MessageFormat {

	var (
		callerSkip int           = 0
		duration   int64         = 0
		errorList  []interface{} = nil
		level      string        = ""
		location   string        = ""
		status     string        = ""
		text       interface{}   = nil
	)

	// Calculate fields.

	timeNow := time.Now().UTC().String()
	callerSkip = messenger.callerSkip
	level = messenger.getLevel(messageNumber)
	id := fmt.Sprintf(messenger.messageIdTemplate, messageNumber)

	textTemplate, ok := messenger.idMessages[messageNumber]
	if ok {
		textRaw := fmt.Sprintf(textTemplate, details...)
		text = strings.Split(textRaw, "%!(")[0]
	}

	statusCandidate, ok := messenger.idStatuses[messageNumber]
	if ok {
		status = statusCandidate
	}

	// TODO: Find status in underlying error.
	// See https://github.com/Senzing/go-logging/blob/48487ee9793e94dac4a3e047635ffd40ff9c454e/messagestatus/messagestatus_senzingapi.go#L29-L59

	// Process Options found in details and filter them out of details.

	filteredDetails := []interface{}{}
	for _, value := range details {
		switch typedValue := value.(type) {
		case *MessageDuration:
			duration = typedValue.Value
		case *MessageId:
			id = typedValue.Value
		case *MessageLevel:
			level = typedValue.Value
		case *MessageLocation:
			location = typedValue.Value
		case *MessageStatus:
			status = typedValue.Value
		case *MessageText:
			text = typedValue.Value
		case *MessageTimestamp:
			timeNow = typedValue.Value.String()
		case *OptionCallerSkip:
			callerSkip = typedValue.Value
		case error:
			errorMessage := typedValue.Error()
			// var priorError interface{}
			if isJson(errorMessage) {
				errorList = append(errorList, jsonAsInterface(errorMessage))
			} else {
				errorList = append(errorList, errorMessage)
			}

			// TODO:
			// messageSplits := strings.Split(errorMessage, "|")
			// for key, value := range SenzingApiErrorsMap {
			// 	if messageSplits[0] == key {
			// 		errorMessageList = append(errorMessageList, value)
			// 	}
			// }
		case time.Duration:
			duration = typedValue.Nanoseconds()
		default:
			filteredDetails = append(filteredDetails, typedValue)
		}
	}

	// Calculate field - location.
	// See https://pkg.go.dev/runtime#Caller

	pc, file, line, ok := runtime.Caller(callerSkip)
	if ok {
		callingFunction := runtime.FuncForPC(pc)
		runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
		functionName := runtimeFunc.ReplaceAllString(callingFunction.Name(), "$1")
		filename := filepath.Base(file)
		location = fmt.Sprintf("In %s() at %s:%d", functionName, filename, line)
	}

	// Compose result.

	result := &MessageFormat{
		Time:     timeNow,
		Level:    level,
		Id:       id,
		Text:     text,
		Status:   status,
		Duration: duration,
		Location: location,
	}

	if len(errorList) > 0 {
		result.Errors = errorList
	}

	if len(filteredDetails) > 0 {
		result.Details = messageDetails(filteredDetails...)
	}

	return result
}

// ----------------------------------------------------------------------------
// Interface methods
// ----------------------------------------------------------------------------

/*
The NewJson method return a JSON string with the elements of the message.

Input
  - messageNumber: A message identifier which indexes into "idMessages".
  - details: Variadic arguments of any type to be added to the message.

Output
  - A JSON string representing the details formatted by the template identified by the messageNumber.
*/
func (messenger *MessengerImpl) NewJson(messageNumber int, details ...interface{}) string {

	messageFormat := messenger.populateStructure(messageNumber, details...)

	// Convert to JSON.

	// Would love to do it this way, but HTML escaping happens.
	// Reported in https://github.com/golang/go/issues/56630
	// result, _ := json.Marshal(messageBuilder)
	// return string(result), err

	// Work-around.

	var resultBytes bytes.Buffer
	enc := json.NewEncoder(&resultBytes)
	enc.SetEscapeHTML(false)
	err := enc.Encode(messageFormat)
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(resultBytes.String())
}

/*
The NewSlog method returns a message and list of Key-Value pairs string with the elements of the message.

Input
  - messageNumber: A message identifier which indexes into "idMessages".
  - details: Variadic arguments of any type to be added to the message.

Output
  - A text message
  - A slice of oscillating key-value pairs.
*/
func (messenger *MessengerImpl) NewSlog(messageNumber int, details ...interface{}) (string, []interface{}) {
	messageFormat := messenger.populateStructure(messageNumber, details...)
	message := ""
	if messageFormat.Text != nil {
		message = messageFormat.Text.(string)
	}
	keys := []string{
		"level",
		"id",
		"status",
		"duration",
		"location",
		"errors",
		"details",
	}
	keyValuePairs := messenger.getKeyValuePairs(messageFormat, keys)
	return message, keyValuePairs
}

/*
The NewSlogLevel method returns a message. an slog level, and a list of Key-Value pairs string with the elements of the message.

Input
  - messageNumber: A message identifier which indexes into "idMessages".
  - details: Variadic arguments of any type to be added to the message.

Output
  - A text message
  - A message level
  - A slice of oscillating key-value pairs.
*/
func (messenger *MessengerImpl) NewSlogLevel(messageNumber int, details ...interface{}) (string, slog.Level, []interface{}) {
	messageFormat := messenger.populateStructure(messageNumber, details...)

	// Create a text message.

	message := ""
	if messageFormat.Text != nil {
		message = messageFormat.Text.(string)
	}

	// Create a slog.Level message level

	slogLevel, ok := TextToLevelMap[messageFormat.Level]
	if !ok {
		slogLevel = LevelPanicSlog
	}

	// Create a slice of oscillating key-value pairs.

	keys := []string{
		"id",
		"status",
		"duration",
		"location",
		"errors",
		"details",
	}
	keyValuePairs := messenger.getKeyValuePairs(messageFormat, keys)
	return message, slogLevel, keyValuePairs
}
