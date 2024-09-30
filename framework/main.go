package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/ANUB1s-xio/TidalFlow/framework/cmd"
)

func main() {

	fmt.Println("Welcome to Tidalflow Framework")
    fmt.Println("Version 1.0.0")


	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("TidalFlow ~ ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")

		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "help":
			cmd.HelpCommand()
			return
		case "eagle":
			//cmd.ExecuteTidalFlow(args[1:])
		case "fly":
			//cmd.SearchCommand(args[1:])
		case "eagle find":
		case "exit":
			fmt.Println("Exiting TidalFlow...")
			return
		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}
