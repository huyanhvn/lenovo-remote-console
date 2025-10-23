package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/huyanhvn/lenovo-remote-console/lenovoconsole"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <BMC_IP> <USERNAME> <PASSWORD> [firefox]")
		fmt.Println("Example: go run main.go 10.145.127.12 USERID PASSW0RD")
		fmt.Println("         go run main.go 10.145.127.12 USERID PASSW0RD firefox")
		fmt.Println("\nNote: Firefox handles BMC connections better than Chrome for this use case")
		os.Exit(1)
	}

	bmcIP := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	// Check if user wants to use Firefox
	useFirefox := false
	if len(os.Args) > 4 && strings.ToLower(os.Args[4]) == "firefox" {
		useFirefox = true
	}

	fmt.Printf("Connecting to XCC at %s...\n", bmcIP)

	// Get RP port
	fmt.Println("Getting remote presence port...")
	rpPort, err := lenovoconsole.GetRPPort(bmcIP, username, password)
	if err != nil {
		fmt.Printf("Warning: Could not get RP port, using default 3900: %v\n", err)
		rpPort = 3900
	}
	fmt.Printf("✓ RP Port: %d\n", rpPort)

	// Create console configuration
	config := lenovoconsole.ConsoleConfig{
		BMCIP:      bmcIP,
		Username:   username,
		Password:   password,
		RPPort:     rpPort,
		UseFirefox: useFirefox,
	}

	// Create and launch console
	console := lenovoconsole.NewConsole(config)

	if err := console.LaunchAndOpen(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print browser-specific instructions
	if useFirefox {
		fmt.Println("\n✓ Firefox launched")
		fmt.Println("  Firefox handles BMC connections well")
	} else {
		fmt.Println("\n✓ Browser launched")
		fmt.Println("  Chrome/Chromium should work well with HTTP connections")
	}

	fmt.Println("\nNote: The browser must be able to reach the XCC at:", bmcIP)
	fmt.Println("\nThis console window will remain active. You can:")
	fmt.Println("  - Launch another instance for a different BMC in a new terminal")
	fmt.Println("  - Press Ctrl+C in this terminal to close this specific console")

	// Keep the program running
	console.WaitForever()
}
