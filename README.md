# Lenovo Remote Console

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/huyanhvn/lenovo-remote-console)](https://goreportcard.com/report/github.com/huyanhvn/lenovo-remote-console)
[![Go Reference](https://pkg.go.dev/badge/github.com/huyanhvn/lenovo-remote-console.svg)](https://pkg.go.dev/github.com/huyanhvn/lenovo-remote-console)

A Go module for connecting to Lenovo XCC (XClarity Controller) remote consoles programmatically. This module handles authentication, WebSocket connections, and serves a web-based KVM console viewer.

## Features

- Connect to Lenovo XCC remote consoles
- Web-based KVM viewer using Lenovo's RPViewer SDK
- Support for multiple simultaneous console sessions
- Automatic certificate handling
- Firefox and Chrome browser support with optimized settings
- Programmatic API for integration into other Go applications

## Installation

### As a Go Module

```bash
go get github.com/huyanhvn/lenovo-remote-console
```

### For CLI Usage

```bash
go install github.com/huyanhvn/lenovo-remote-console/cmd/lenovo-console@latest
```

## Usage

### Command Line Interface

```bash
# Basic usage
lenovo-console <BMC_IP> <USERNAME> <PASSWORD>

# Use Firefox (recommended for better certificate handling)
lenovo-console 10.145.127.12 admin password firefox

# Or run directly with go run
go run main.go 10.145.127.12 admin password
```

### As a Go Module

```go
package main

import (
    "log"
    "github.com/huyanhvn/lenovo-remote-console/lenovoconsole"
)

func main() {
    // Create console configuration
    config := lenovoconsole.ConsoleConfig{
        BMCIP:      "10.145.127.12",
        Username:   "admin",
        Password:   "password",
        UseFirefox: true,  // Optional: prefer Firefox
        ServerPort: 8443,  // Optional: specify port (0 for auto)
    }

    // Create and launch console
    console := lenovoconsole.NewConsole(config)
    
    // Simple one-line launch
    if err := console.LaunchAndOpen(); err != nil {
        log.Fatal(err)
    }

    // Keep console running
    console.WaitForever()
}
```

### Advanced Usage

```go
// Step-by-step control
console := lenovoconsole.NewConsole(config)

// Initialize (generates HTML, finds port)
if err := console.Initialize(); err != nil {
    log.Fatal(err)
}

// Start the server
if err := console.Start(); err != nil {
    log.Fatal(err)
}

// Get the console URL
url := console.GetURL()
fmt.Printf("Console available at: %s\n", url)

// Open in browser when ready
if err := console.OpenInBrowser(); err != nil {
    log.Fatal(err)
}

// Later... stop the console
console.Stop()
```

### Multiple Consoles

```go
// Launch consoles for multiple BMCs
bmcs := []lenovoconsole.ConsoleConfig{
    {BMCIP: "10.145.127.12", Username: "admin", Password: "pass1"},
    {BMCIP: "10.145.127.13", Username: "admin", Password: "pass2"},
}

for _, config := range bmcs {
    console := lenovoconsole.NewConsole(config)
    go func(c *lenovoconsole.Console) {
        if err := c.LaunchAndOpen(); err != nil {
            log.Printf("Failed: %v", err)
        }
    }(console)
}

select {} // Keep running
```

## API Reference

### Types

#### `ConsoleConfig`
Configuration for a remote console session:
- `BMCIP`: IP address of the BMC/XCC
- `Username`: Authentication username
- `Password`: Authentication password
- `RPPort`: Remote Presence port (default: 3900)
- `UseFirefox`: Prefer Firefox browser
- `ServerPort`: Local server port (0 for auto-assign)

#### `Console`
Main console object with methods:
- `NewConsole(config)`: Create new console instance
- `Initialize()`: Prepare console for launch
- `Start()`: Start the HTTPS server
- `Stop()`: Stop the console server
- `OpenInBrowser()`: Open console in browser
- `LaunchAndOpen()`: Combined Initialize + Start + OpenInBrowser
- `GetURL()`: Get the console URL
- `GetPort()`: Get the server port
- `WaitForever()`: Block forever (keeps console running)

### Functions

#### `GetRPPort(bmcIP, username, password)`
Query the XCC for the Remote Presence port. Returns port number or 3900 as default.

## Certificate Requirements

The console requires HTTPS with proper certificates. You'll need:

1. `server.crt` and `server.key` files in the working directory for the local HTTPS server
2. Accept the self-signed certificate warnings for both localhost and the BMC

### Generating Self-Signed Certificates

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
  -days 365 -nodes -subj "/CN=localhost"
```

## Browser Compatibility

- **Firefox** (Recommended): Better handling of self-signed certificates
- **Chrome/Chromium**: Works with special flags to bypass certificate warnings
- **Safari**: Should work but may require manual certificate acceptance

## Examples

See the `examples/` directory for more usage examples:
- `multiple_consoles.go`: Managing multiple console sessions
- Programmatic control examples
- Integration patterns

## Requirements

- Go 1.21 or later
- Network access to Lenovo XCC/BMC
- Modern web browser (Firefox recommended)
- HTTPS certificates (self-signed acceptable)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Security

If you discover a security vulnerability within this project, please create an issue or contact the maintainers directly. All security vulnerabilities will be promptly addressed.

## Acknowledgments

- Lenovo for the XCC RPViewer SDK
- The Go community for excellent tools and libraries
- All contributors who help improve this project

## Troubleshooting

### Certificate Issues
If you see "Certificate not verified" errors:
1. Accept the localhost certificate first
2. Open `https://<BMC_IP>:3900` in a new tab
3. Accept the BMC certificate
4. Return to console tab and click "Retry Connection"

### WebSocket Errors
- Try using Firefox instead of Chrome
- Ensure the BMC firmware is up to date
- Check network connectivity to the BMC

### Port Conflicts
- The module automatically finds available ports
- You can specify a fixed port in `ConsoleConfig.ServerPort`
