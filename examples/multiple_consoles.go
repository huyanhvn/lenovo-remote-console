// Example demonstrating how to use the lenovoconsole module to manage multiple console sessions
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/huyanhvn/lenovo-remote-console/lenovoconsole"
)

func main() {
	// Example 1: Simple console launch
	simpleConsoleExample()

	// Example 2: Multiple consoles
	// multipleConsolesExample()

	// Example 3: Programmatic control
	// programmaticExample()
}

// simpleConsoleExample demonstrates basic console usage
func simpleConsoleExample() {
	config := lenovoconsole.ConsoleConfig{
		BMCIP:      "10.145.127.12",
		Username:   "USERID",
		Password:   "PASSW0RD",
		UseFirefox: true, // Prefer Firefox for better certificate handling
	}

	console := lenovoconsole.NewConsole(config)

	// Launch and open in one go
	if err := console.LaunchAndOpen(); err != nil {
		log.Fatalf("Failed to launch console: %v", err)
	}

	fmt.Printf("Console running at: %s\n", console.GetURL())

	// Keep running
	console.WaitForever()
}

// multipleConsolesExample demonstrates launching multiple consoles for different BMCs
func multipleConsolesExample() {
	bmcs := []struct {
		IP       string
		Username string
		Password string
	}{
		{"10.145.127.12", "admin", "password1"},
		{"10.145.127.13", "admin", "password2"},
		{"10.145.127.14", "admin", "password3"},
	}

	consoles := make([]*lenovoconsole.Console, 0, len(bmcs))

	// Launch all consoles
	for _, bmc := range bmcs {
		config := lenovoconsole.ConsoleConfig{
			BMCIP:    bmc.IP,
			Username: bmc.Username,
			Password: bmc.Password,
		}

		console := lenovoconsole.NewConsole(config)

		// Initialize and start server without opening browser
		if err := console.Initialize(); err != nil {
			log.Printf("Failed to initialize console for %s: %v", bmc.IP, err)
			continue
		}

		if err := console.Start(); err != nil {
			log.Printf("Failed to start console for %s: %v", bmc.IP, err)
			continue
		}

		consoles = append(consoles, console)
		fmt.Printf("Console for BMC %s running at: %s\n", bmc.IP, console.GetURL())
	}

	// Open all consoles in browser after a delay
	time.Sleep(2 * time.Second)
	for _, console := range consoles {
		if err := console.OpenInBrowser(); err != nil {
			log.Printf("Failed to open browser: %v", err)
		}
		time.Sleep(500 * time.Millisecond) // Small delay between opening tabs
	}

	fmt.Println("\nAll consoles launched. Press Ctrl+C to exit.")

	// Keep all consoles running
	select {}
}

// programmaticExample demonstrates more fine-grained control
func programmaticExample() {
	config := lenovoconsole.ConsoleConfig{
		BMCIP:      "10.145.127.12",
		Username:   "admin",
		Password:   "password",
		ServerPort: 8443, // Specify a fixed port
	}

	console := lenovoconsole.NewConsole(config)

	// Step 1: Initialize (generates HTML, finds port if needed)
	fmt.Println("Initializing console...")
	if err := console.Initialize(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf("Console will run on port: %d\n", console.GetPort())

	// Step 2: Start the server
	fmt.Println("Starting server...")
	if err := console.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	fmt.Printf("Server running at: %s\n", console.GetURL())

	// Step 3: Wait a bit, then open browser
	fmt.Println("Waiting 3 seconds before opening browser...")
	time.Sleep(3 * time.Second)

	fmt.Println("Opening browser...")
	if err := console.OpenInBrowser(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}

	// Step 4: Run for some time
	fmt.Println("Console will run for 1 minute...")
	time.Sleep(1 * time.Minute)

	// Step 5: Shutdown
	fmt.Println("Shutting down console...")
	if err := console.Stop(); err != nil {
		log.Printf("Failed to stop console: %v", err)
	}

	fmt.Println("Console stopped.")
}
