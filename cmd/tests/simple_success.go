package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "5678")
	fmt.Fprint(os.Stderr, "qwer")
}
