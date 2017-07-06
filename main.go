package main

import (
	"flag"
	"fmt"
)

func main() {
	/*ipPtr := flag.String("ip", "10.0.0.1", "a string")
	logPtr := flag.String("log", "admin", "a string")

	numbPtr := flag.Int("numb", 42, "an int")
	newPtr := flag.Bool("new", true, "a bool")

	var svar string
	flag.StringVar(&svar, "svar", "bar", "a string var")

	flag.Parse()

	fmt.Println("ip:", *ipPtr)
	fmt.Println("log:", *logPtr)
	fmt.Println("numb:", *numbPtr)
	fmt.Println("new:", *newPtr)
	fmt.Println("svar:", svar)
	fmt.Println("tail:", flag.Args())

	//defer recover()*/
	helpArg := flag.Bool("help", false, "a boolean")

	flag.Parse()

	if *helpArg == true {
		fmt.Println("\nAllowed commands:")
	}
}
