// Command dupkey demonstrates encoding/json (v1) behavior when two struct
// fields map to the same JSON key via their tags.
package main

import (
	"encoding/json"
	"fmt"
)

type S struct {
	Alpha string `json:"key"`
	Beta  string `json:"Key"`
}

func main() {
	s := S{Alpha: "one", Beta: "two"}
	b, err := json.Marshal(s)
	fmt.Printf("result: %q\n", b)
	fmt.Printf("err:    %v\n", err)
}
