// Package lenovoconsole provides a Go client library for connecting to Lenovo XCC remote consoles.
// It handles authentication, establishes WebSocket connections, and serves a web-based console viewer.
package lenovoconsole

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// TokenResponse represents the authentication token response from XCC
type TokenResponse struct {
	Token string `json:"Token"`
}

// Credentials contains the authentication credentials for XCC
type Credentials struct {
	Username string
	Password string
}

// ConsoleConfig contains configuration for a remote console session
type ConsoleConfig struct {
	BMCIP      string // IP address of the BMC/XCC
	Username   string // Username for authentication
	Password   string // Password for authentication
	RPPort     int    // Remote Presence port (default: 3900)
	UseFirefox bool   // Whether to prefer Firefox browser
	ServerPort int    // Local server port (0 for auto-assign)
}

// Console represents a remote console session
type Console struct {
	config      ConsoleConfig
	serverPort  int
	server      *http.Server
	consoleHTML string
	mux         *http.ServeMux
}

// NewConsole creates a new Console instance with the given configuration
func NewConsole(config ConsoleConfig) *Console {
	return &Console{
		config: config,
		mux:    http.NewServeMux(),
	}
}

// GetRPPort queries the XCC for the Remote Presence port
// Returns the port number or 3900 as default if query fails
func GetRPPort(bmcIP, username, password string) (int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("https://%s/api/providers/rp_port", bmcIP)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 3900, nil // default port
	}

	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return 3900, nil // default port
	}
	defer resp.Body.Close()

	var result struct {
		Port int `json:"port"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Port > 0 {
		return result.Port, nil
	}
	return 3900, nil // default port
}

// Initialize prepares the console for launch
func (c *Console) Initialize() error {
	// Get RP port if not set
	if c.config.RPPort == 0 {
		port, err := GetRPPort(c.config.BMCIP, c.config.Username, c.config.Password)
		if err != nil {
			return fmt.Errorf("failed to get RP port: %v", err)
		}
		c.config.RPPort = port
	}

	// Generate HTML
	if err := c.generateHTML(); err != nil {
		return fmt.Errorf("failed to generate HTML: %v", err)
	}

	// Find available port if not specified
	if c.config.ServerPort == 0 {
		port, err := findAvailablePort()
		if err != nil {
			return fmt.Errorf("failed to find available port: %v", err)
		}
		c.serverPort = port
	} else {
		c.serverPort = c.config.ServerPort
	}

	// Setup HTTP handlers
	c.setupHandlers()

	return nil
}

// Start begins serving the console on the configured port
// This method does not block
func (c *Console) Start() error {
	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.serverPort),
		Handler: c.mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	go func() {
		if err := c.server.ListenAndServeTLS("server.crt", "server.key"); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTPS server error for BMC %s: %v\n", c.config.BMCIP, err)
		}
	}()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	return nil
}

// Stop gracefully shuts down the console server
func (c *Console) Stop() error {
	if c.server != nil {
		return c.server.Close()
	}
	return nil
}

// GetURL returns the URL to access the console
func (c *Console) GetURL() string {
	return fmt.Sprintf("https://localhost:%d", c.serverPort)
}

// GetPort returns the local server port
func (c *Console) GetPort() int {
	return c.serverPort
}

// OpenInBrowser opens the console in the default or specified browser
func (c *Console) OpenInBrowser() error {
	consoleURL := c.GetURL()

	cmd, err := c.getBrowserCommand(consoleURL)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %v", err)
	}

	return nil
}

// LaunchAndOpen initializes, starts the server, and opens the console in a browser
// This is a convenience method that combines Initialize, Start, and OpenInBrowser
func (c *Console) LaunchAndOpen() error {
	if err := c.Initialize(); err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		return err
	}

	if err := c.OpenInBrowser(); err != nil {
		return err
	}

	fmt.Println("\n✓ Console launched in browser")
	fmt.Printf("Console URL: %s\n", c.GetURL())
	fmt.Printf("BMC IP: %s (Port: %d)\n", c.config.BMCIP, c.config.RPPort)
	fmt.Println("✓ This console instance is running independently")

	return nil
}

// WaitForever blocks forever, keeping the console server running
func (c *Console) WaitForever() {
	select {}
}

// generateHTML creates the HTML content for the console viewer
func (c *Console) generateHTML() error {
	tmpl, err := template.New("console").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var buf strings.Builder
	data := struct {
		BMCIP       string
		RPPort      int
		BMCUsername string
		BMCPassword string
	}{
		BMCIP:       c.config.BMCIP,
		RPPort:      c.config.RPPort,
		BMCUsername: c.config.Username,
		BMCPassword: c.config.Password,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	c.consoleHTML = buf.String()
	return nil
}

// setupHandlers configures the HTTP handlers for the console
func (c *Console) setupHandlers() {
	// Main console handler
	c.mux.HandleFunc("/", c.consoleHandler)
	c.mux.HandleFunc("/cert.pem", certHandler)

	// Proxy handlers for SDK files
	proxyHandler := c.proxySDKHandler()
	c.mux.HandleFunc("/SDK_Pilot4/", proxyHandler)
	c.mux.HandleFunc("/offscreenworker.js", proxyHandler)
	c.mux.HandleFunc("/mouseworker.js", proxyHandler)
	c.mux.HandleFunc("/utility.js", proxyHandler)
	c.mux.HandleFunc("/mediaTypes.js", proxyHandler)
	c.mux.HandleFunc("/rphandlers.js", proxyHandler)
	c.mux.HandleFunc("/websockethandler.js", proxyHandler)
	c.mux.HandleFunc("/virtualkeyboard.js", proxyHandler)
	c.mux.HandleFunc("/mediaworkerhandler.js", proxyHandler)
}

// consoleHandler serves the main console HTML
func (c *Console) consoleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(c.consoleHTML))
}

// proxySDKHandler creates a handler to proxy SDK files from BMC
func (c *Console) proxySDKHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		bmcURL := fmt.Sprintf("https://%s%s", c.config.BMCIP, r.URL.Path)

		req, err := http.NewRequest(r.Method, bmcURL, nil)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch from BMC", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}

		w.WriteHeader(resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	}
}

// getBrowserCommand returns the appropriate command to open the browser
func (c *Console) getBrowserCommand(url string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		return c.getWindowsBrowserCommand(url)
	case "darwin":
		return c.getDarwinBrowserCommand(url)
	case "linux":
		return c.getLinuxBrowserCommand(url)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (c *Console) getWindowsBrowserCommand(url string) (*exec.Cmd, error) {
	if c.config.UseFirefox {
		firefoxPaths := []string{
			"C:\\Program Files\\Mozilla Firefox\\firefox.exe",
			"C:\\Program Files (x86)\\Mozilla Firefox\\firefox.exe",
		}
		for _, ffPath := range firefoxPaths {
			if _, err := os.Stat(ffPath); err == nil {
				return exec.Command(ffPath, url), nil
			}
		}
	}

	// Try Chrome with flags
	chromePath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	if _, err := os.Stat(chromePath); err == nil {
		return exec.Command(chromePath,
			"--ignore-certificate-errors",
			"--test-type",
			"--allow-insecure-localhost",
			"--disable-popup-blocking",
			"--disable-blink-features=AutomationControlled",
			"--disable-session-crashed-bubble",
			"--disable-infobars",
			"--no-first-run",
			"--no-default-browser-check",
			"--user-data-dir="+os.TempDir()+"/chrome-temp-profile",
			url), nil
	}

	// Fallback to default browser
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url), nil
}

func (c *Console) getDarwinBrowserCommand(url string) (*exec.Cmd, error) {
	if c.config.UseFirefox {
		firefoxPath := "/Applications/Firefox.app/Contents/MacOS/firefox"
		if _, err := os.Stat(firefoxPath); err == nil {
			return exec.Command(firefoxPath, url), nil
		}
	}

	// Try Chrome with flags
	chromePath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if _, err := os.Stat(chromePath); err == nil {
		return exec.Command(chromePath,
			"--ignore-certificate-errors",
			"--test-type",
			"--allow-insecure-localhost",
			"--disable-popup-blocking",
			"--disable-blink-features=AutomationControlled",
			"--disable-session-crashed-bubble",
			"--disable-infobars",
			"--no-first-run",
			"--no-default-browser-check",
			"--user-data-dir="+os.TempDir()+"/chrome-temp-profile",
			url), nil
	}

	// Fallback to default browser
	return exec.Command("open", url), nil
}

func (c *Console) getLinuxBrowserCommand(url string) (*exec.Cmd, error) {
	if c.config.UseFirefox {
		firefoxPaths := []string{
			"/usr/bin/firefox",
			"/usr/local/bin/firefox",
			"/snap/bin/firefox",
		}
		for _, ffPath := range firefoxPaths {
			if _, err := os.Stat(ffPath); err == nil {
				return exec.Command(ffPath, url), nil
			}
		}
	}

	// Try Chrome/Chromium with flags
	chromePaths := []string{
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
	}
	for _, chromePath := range chromePaths {
		if _, err := os.Stat(chromePath); err == nil {
			return exec.Command(chromePath,
				"--ignore-certificate-errors",
				"--test-type",
				"--allow-insecure-localhost",
				"--disable-popup-blocking",
				"--disable-blink-features=AutomationControlled",
				"--disable-session-crashed-bubble",
				"--disable-infobars",
				"--no-first-run",
				"--no-default-browser-check",
				"--user-data-dir="+os.TempDir()+"/chrome-temp-profile",
				url), nil
		}
	}

	// Fallback to default browser
	return exec.Command("xdg-open", url), nil
}

// certHandler serves a dummy certificate page for RPViewer
func certHandler(w http.ResponseWriter, r *http.Request) {
	certHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Certificate Acceptance</title>
    <script>
        function acceptCertificate() {
            if (window.opener && !window.opener.closed) {
                try {
                    window.opener.postMessage({
                        type: 'certificate',
                        action: 'accept',
                        accepted: true
                    }, '*');
                    window.opener.postMessage('CERT_ACCEPTED', '*');
                } catch(e) {
                    console.log('Could not post message to opener:', e);
                }
            }
            
            try {
                localStorage.setItem('rpviewer_cert_accepted', 'true');
            } catch(e) {
                console.log('Could not set localStorage:', e);
            }
            
            if (window.opener && window.opener.rpCertAccepted) {
                try {
                    window.opener.rpCertAccepted();
                } catch(e) {
                    console.log('Could not call rpCertAccepted:', e);
                }
            }
            
            setTimeout(function() {
                window.close();
            }, 100);
        }
        
        window.onload = acceptCertificate;
        acceptCertificate();
    </script>
</head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
    <h3>Certificate Handler</h3>
    <p>The SSL certificate has been accepted. This window will close automatically.</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(certHTML))
}

// findAvailablePort finds an available port on the system
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
