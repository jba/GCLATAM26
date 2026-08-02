// Command dupopt fires the RFC 7493 duplicate-name check using a map whose
// key type marshals two distinct Go keys to the SAME JSON name.
package main

import (
	"fmt"

	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	json "encoding/json/v2"
)

// K is a map key type. Every value marshals to the same JSON object name,
// so a map with two distinct K keys produces duplicate names.
type K int

func (K) MarshalText() ([]byte, error) { return []byte("dup"), nil }

func main() {
	m := map[K]int{1: 10, 2: 20}
	b, err := json.Marshal(m,
		jsonv1.DefaultOptionsV1(),
		jsontext.AllowDuplicateNames(false),
	)
	fmt.Printf("result: %q\n", b)
	fmt.Printf("err:    %v\n", err)
}
