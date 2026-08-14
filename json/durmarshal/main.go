// Command durmarshal registers a custom json/v2 marshaler for
// time.Duration that encodes durations as their human-readable string
// form (e.g. "1h30m0s") instead of the default integer nanoseconds.
package main

import (
	"fmt"
	"time"

	json "encoding/json/v2"
)

type Task struct {
	Name    string        `json:"name"`
	Timeout time.Duration `json:"timeout"`
	Elapsed time.Duration `json:"elapsed"`
}

func main() {
	// A type-specific marshaler for time.Duration. It must produce exactly
	// one JSON value; here a JSON string built from Duration.String.
	durMarshaler := json.MarshalFunc(func(d time.Duration) ([]byte, error) {
		return []byte(`"` + d.String() + `"`), nil
	})

	task := Task{
		Name:    "backup",
		Timeout: 90 * time.Minute,
		Elapsed: 1500 * time.Millisecond,
	}

	// Without the custom marshaler, json/v2 has no default representation
	// for time.Duration and reports an error (see go.dev/issue/71631).
	if _, err := json.Marshal(task); err != nil {
		fmt.Printf("default: error: %v\n", err)
	}

	// With the custom marshaler registered via WithMarshalers.
	customOut, err := json.Marshal(task, json.WithMarshalers(durMarshaler))
	if err != nil {
		panic(err)
	}
	fmt.Printf("custom:  %s\n", customOut)
}
