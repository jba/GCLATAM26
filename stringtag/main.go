// Command jsonstringtag demonstrates the ",string" tag option under
// encoding/json/v2 on a variety of field types. Each field is marshaled
// independently so the result (or error) for every field is shown.
package main

import (
	"fmt"

	json "encoding/json/v2"
)

type (
	Int struct {
		V int `json:"v,string"`
	}
	Float struct {
		V float64 `json:"v,string"`
	}
	IntSlice struct {
		V []int `json:"v,string"`
	}
	FloatSlice struct {
		V []float64 `json:"v,string"`
	}
	IntMap struct {
		V map[string]int `json:"v,string"`
	}
	FloatMap struct {
		V map[string]float64 `json:"v,string"`
	}
	AnyMap struct {
		V map[string]any `json:"v,string"`
	}
	String struct {
		V string `json:"v,string"`
	}
	Bool struct {
		V bool `json:"v,string"`
	}
)

func show(name string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("%-11s error: %v\n", name, err)
		return
	}
	fmt.Printf("%-11s %s\n", name, b)
}

func main() {
	show("int", Int{1})
	show("float", Float{2.5})
	show("intSlice", IntSlice{[]int{3, 4}})
	show("floatSlice", FloatSlice{[]float64{5.5, 6.5}})
	show("intMap", IntMap{map[string]int{"a": 7}})
	show("floatMap", FloatMap{map[string]float64{"b": 8.5}})
	show("anyMap", AnyMap{map[string]any{"c": 9, "d": []int{10, 11}}})
	show("string", String{"hello"})
	show("bool", Bool{true})
}
