package main

import (
	"fmt"
	//"github.com/sirupsen/logrus"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "1234")
	fmt.Fprint(os.Stderr, "asdf")
}
