package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	
	"golang.org/x/term"
)

var (
	startTime = time.Now()
	daemon    *Daemon
)

func main() {
	var listener net.Listener
	var err error
	var port string

	// Try to listen on port 42 first
	listener, err = net.Listen("tcp", "127.0.0.1:42")
	if err != nil {
		// Check if it's specifically a permission error
		if strings.Contains(err.Error(), "permission denied") {
			// Check if running non-interactively (e.g., with nohup)
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				// Non-interactive mode - just fall back to 4242
				log.Println("🔐 Port 42 requires elevated permissions. Falling back to port 4242...")
				listener, err = net.Listen("tcp", "127.0.0.1:4242")
				if err != nil {
					log.Fatal("Failed to open Port 4242:", err)
				}
				port = "4242"
				log.Println("🐬 Swimming on port 4242...")
			} else {
				// Interactive mode - show prompt
				fmt.Println("🔐 Port 42 requires elevated permissions.")
				fmt.Println("🐬 The dolphins need permission to swim in the sacred waters of Port 42.")
				fmt.Println("\nOptions:")
				fmt.Println("1. Run with sudo: sudo port42d")
				fmt.Println("2. Use port 4242 instead (no permissions needed)")
				fmt.Print("\nPress Enter to use port 4242, or Ctrl+C to exit and run with sudo: ")
				
				// Wait for user input
				fmt.Scanln()
				
				// Try port 4242
				listener, err = net.Listen("tcp", "127.0.0.1:4242")
				if err != nil {
					log.Fatal("Failed to open Port 4242:", err)
				}
				port = "4242"
				log.Println("🐬 Swimming on port 4242...")
			}
		} else {
			// Some other error (like port already in use)
			log.Fatal("Failed to open Port 42:", err)
		}
	} else {
		port = "42"
		log.Println("🐬 Port 42 is open. The dolphins are listening...")
	}
	
	// Log the actual port we're using
	log.Printf("◊ Listening on localhost:%s", port)
	
	// Debug environment - check PORT42_ANTHROPIC_API_KEY first, then ANTHROPIC_API_KEY
	apiKey := os.Getenv("PORT42_ANTHROPIC_API_KEY")
	if apiKey != "" {
		log.Printf("🔑 Using PORT42_ANTHROPIC_API_KEY (length: %d)", len(apiKey))
	} else {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey != "" {
			log.Printf("🔑 Using ANTHROPIC_API_KEY (length: %d)", len(apiKey))
		} else {
			log.Println("")
			log.Println("⚠️ ⚠️ ⚠️  WARNING: NO API KEY FOUND ⚠️ ⚠️ ⚠️")
			log.Println("")
			log.Println("Port 42 AI features will NOT work without an API key!")
			log.Println("")
			log.Println("To fix this:")
			log.Println("  1. Set PORT42_ANTHROPIC_API_KEY or ANTHROPIC_API_KEY")
			log.Println("  2. Restart the daemon")
			log.Println("")
			log.Println("Example:")
			log.Println("  export PORT42_ANTHROPIC_API_KEY='your-key-here'")
			log.Println("  port42 daemon start")
			log.Println("")
		}
	}
	
	// Check if running under sudo
	if apiKey == "" {
		log.Printf("🔍 Running as user: %s (UID: %d)", os.Getenv("USER"), os.Getuid())
		log.Printf("🔍 SUDO_USER: %s", os.Getenv("SUDO_USER"))
		log.Printf("🔍 HOME: %s", os.Getenv("HOME"))
		
		// List all env vars starting with ANTHRO
		log.Println("🔍 Environment variables containing 'ANTHRO':")
		for _, env := range os.Environ() {
			if strings.Contains(env, "ANTHRO") {
				log.Printf("   %s", env)
			}
		}
	}

	// Load agent configuration
	if err := LoadAgentConfig(); err != nil {
		log.Printf("⚠️  Failed to load agent config: %v", err)
	}

	// Create daemon
	daemon = NewDaemon(listener, port)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Start daemon in goroutine
	go daemon.Start()

	// Wait for shutdown signal
	<-sigChan
	log.Println("\n🐬 The dolphins are returning to the depths...")
	
	// Graceful shutdown
	daemon.Shutdown()
}

