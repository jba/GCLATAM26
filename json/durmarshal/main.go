// Command durmarshal shows three ways to encode time.Duration values as
// their human-readable string form (e.g. "1h30m0s") under json/v2, which
// otherwise has no default representation for durations:
//
//   - default:  json.Marshal fails (no representation for time.Duration).
//   - custom:   a json.MarshalFunc registered via json.WithMarshalers.
//   - marshalto: a json.MarshalToFunc (MarshalJSONTo style) registered
//     via json.WithMarshalers, writing directly to the encoder.
package main

import (
	"fmt"
	"time"

	"encoding/json/jsontext"
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

	// Third approach: also a custom marshaler (not a method on any type),
	// but written in the MarshalJSONTo style using json.MarshalToFunc. The
	// function writes directly to the jsontext.Encoder instead of returning
	// bytes, which is more performant.
	durMarshalerTo := json.MarshalToFunc(func(enc *jsontext.Encoder, d time.Duration) error {
		return enc.WriteToken(jsontext.String(d.String()))
	})
	methodOut, err := json.Marshal(task, json.WithMarshalers(durMarshalerTo))
	if err != nil {
		panic(err)
	}
	fmt.Printf("marshalto: %s\n", methodOut)
}
