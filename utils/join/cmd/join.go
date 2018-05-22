package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		return
	}
	bar := os.Args[1]
	raw := os.Args[2:]
	for i := range raw {
		raw[i] = fmt.Sprintf("%s%s%s", bar, raw[i], bar)
	}
	fmt.Println(strings.Join(raw, ","))
}
