package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/ANUB1s-xio/EagleDeployment/compiled-framework/cmd"
)

func main() {

	fmt.Println("Welcome to EagleDeployment")
    fmt.Println("Version 1.0.0")


	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("EagleDeployment ~ ")
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
		case "cat":
		case "exit":
			fmt.Println("Exiting EagleDeployment...")
			return
		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}
