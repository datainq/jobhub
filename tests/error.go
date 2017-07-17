package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "1234")
	fmt.Fprint(os.Stderr, "asdf")
}
