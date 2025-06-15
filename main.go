package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println("Hello, world!")
}

func Marshal(v any) string {
	ret, _ := json.Marshal(v)
	return string(ret)
}
