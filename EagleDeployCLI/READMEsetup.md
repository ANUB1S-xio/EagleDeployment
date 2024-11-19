
# Set Up Go Environment:

## Install Go:
- Download and install Go from the official Go website.
- Verify the installation by running `go version` in your terminal or command prompt.

## Create Your Go Application:

- Create a new directory for your project:
  ```bash
  mkdir eagle
  cd eagle
  ```
- Create a Go file (e.g., main.go) and write your code.

## Initialize Go Module (if needed):

Run the following command to create a go.mod file for dependency management:
```bash
go mod init eagle
```

## Build the Application:

### For the Current OS:

Navigate to the directory containing your Go file and run:
```bash
go build main.go
```
This will create an executable file named eagle (or eagle.exe depending on the OS).

### For Cross-Compilation:

To build for a different OS, set the GOOS and GOARCH environment variables:

#### Windows from Linux/macOS:
```bash
GOOS=windows GOARCH=amd64 go build -o eagle.exe main.go
```

#### Linux from Windows:
```bash
GOOS=linux GOARCH=amd64 go build -o eagle main.go
```

## Run the Application:

### On the Same OS:

Run the executable directly:
```bash
./eagle    # On Linux/macOS
eagle.exe  # On Windows
```

### On a Different OS:

After cross-compilation, transfer the binary to the target machine.
Then run it as mentioned above for the respective OS.

## Check for Dependencies:

If your application uses external libraries, ensure they are available on the target OS.
Use `go mod tidy` to ensure all dependencies are correctly set up.

## Handle OS-Specific Considerations:

Adjust file paths and I/O operations to be compatible with the operating system.
If necessary, use Docker for a consistent environment across different platforms.

# Summary of Running Your Go Application

To run your Go application, follow these steps in your terminal based on your directory structure:

## 1. Open Your Terminal
Ensure you're in the directory where your Go files are located. Navigate to your project directory if necessary:
```bash
cd /path/to/your/testDeploy
```

## 2. Ensure Go is Installed
Before running your Go application, check if Go is installed by running:
```bash
go version
```

## 3. Build the Go Application
You can build your application using the following command:
```bash
go build main.go
```
This command creates an executable file named `main` (or `main.exe` on Windows) in the same directory.

## 4. Run the Application
After building, run the executable directly:
```bash
./main    # On Linux/macOS
main.exe  # On Windows
```

## 5. Run Without Building
Alternatively, you can run your Go application directly using:
```bash
go run main.go
```

## 6. Check for Dependencies
If you have dependencies defined in `go.mod`, ensure they are downloaded by running:
```bash
go mod tidy
```

## 7. Check Logs/Output
If your application produces output or logs, check the terminal for any messages or results.

## 8. Command-Line Arguments
If your application requires specific arguments or flags, include them after `./main` or `go run main.go`.

## Locating the Binary Executable
- After running `go build main.go`, the binary executable is created in the same directory. 
- To check if the binary has already been created, use:
```bash
ls
```
Look for an executable file named `main` (or `main.exe` on Windows).

## Running the Binary
If the binary exists, run it directly:
```bash
./main    # On Linux/macOS
main.exe  # On Windows
```

If the binary isn't present or if you made changes to your source code, run `go build` again to create the most current version. This ensures you are executing the latest build of your application.

## Checking the Binary Location
1. Navigate to the project directory:
```bash
cd /path/to/your/EagleDeploy
```
2. List all files, including binaries, using:
```bash
ls -l
```
3. Confirm the file type using:
```bash
file main
```
If you used `go install`, the binary would typically be located in your `$GOPATH/bin` directory. If you don't see the binary and used `go build`, check the output for error messages that may indicate why it wasn't created.
