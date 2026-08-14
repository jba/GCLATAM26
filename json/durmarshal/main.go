// Command durmarshal shows three ways to encode time.Duration values as
// their human-readable string form (e.g. "1h30m0s") under json/v2, which
// otherwise has no default representation for durations:
//
//   - default:  json.Marshal fails (no representation for time.Duration).
//   - custom:   a json.MarshalFunc registered via json.WithMarshalers.
//   - method:   a wrapper type implementing MarshalerTo (MarshalJSONTo).
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

// Duration is a time.Duration wrapper that implements json/v2's
// MarshalerTo interface (the MarshalJSONTo method), writing itself as a
// JSON string. This is the recommended, more performant alternative to
// registering a MarshalFunc: the behavior travels with the type.
type Duration time.Duration

func (d Duration) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String(time.Duration(d).String()))
}

// StructTask mirrors Task but uses the self-marshaling Duration type.
type StructTask struct {
	Name    string   `json:"name"`
	Timeout Duration `json:"timeout"`
	Elapsed Duration `json:"elapsed"`
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

	// Third approach: a wrapper type that implements MarshalerTo
	// (MarshalJSONTo). No options needed at the call site — the type
	// marshals itself.
	structTask := StructTask{
		Name:    "backup",
		Timeout: Duration(90 * time.Minute),
		Elapsed: Duration(1500 * time.Millisecond),
	}
	methodOut, err := json.Marshal(structTask)
	if err != nil {
		panic(err)
	}
	fmt.Printf("method:  %s\n", methodOut)
}
