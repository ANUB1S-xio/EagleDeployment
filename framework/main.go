package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/ANUB1s-xio/TidalFlow/framework/cmd"
)

func main() {
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
		case "Tide":
			//cmd.ExecuteTidalFlow(args[1:])
		case "swim":
			//cmd.SearchCommand(args[1:])
		case "exit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}
