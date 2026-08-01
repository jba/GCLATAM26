// Command jsonstringtag demonstrates the ",string" tag option under
// encoding/json/v2 on a map[string]any with heterogeneous nested values.
package main

import (
	"fmt"

	json "encoding/json/v2"
)

type S struct {
	M map[string]any `json:"m,string"`
}

func main() {
	s := S{
		M: map[string]any{
			"a": int64(1),
			"b": []int64{2, 3},
			"c": map[string]any{
				"x": int64(4),
				"y": []int64{5, 6},
			},
		},
	}
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(b))
}
