package EagleDeployment

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)
func main() {
	//fmt.Println("Welcome to EagleDeployment")
    //fmt.Println("Version 1.0.0")
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
			//cmd.HelpCommand()
			return
		case "eagle":
			//cmd.EagleCommand(args[1:])
		case "soar":
			//cmd.SoarCommand(args[1:])
		case "display":
			//cmd.DisplayCommand(args[1:])
		case "exit":
			fmt.Println("Exiting EagleDeployment...")
			return
		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}
/*----------------------------- case logic ------------------------------------*/
func eagleCommand(input string) {
    // Example: handle file execution logic
    fmt.Println("Executing file:", input)
}
func soarCommand(input string) {
    // Example: handle file path search
    fmt.Println("Searching file:", input)
}
func displayCommand(input string) {
    // Example: handle file display logic
    fmt.Println("Displaying file contents for:", input)
}
// HelpCommand prints the available commands
func HelpCommand() {
	fmt.Println("Available commands:")
	fmt.Println("  eagle [file]              - Execute file")
	fmt.Println("  soar [file/keyword]       - Search file")
    fmt.Println("  display [file]            - Display file contents")
	fmt.Println("  exit                      - Exit the framework")
}

