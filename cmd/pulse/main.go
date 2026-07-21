package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/valtors/pulse/internal/server"
)

const version = "0.1.0"

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "version" || os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println(version)
		return
	}

	if len(os.Args) >= 2 && (os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help") {
		usage()
		return
	}

	fs := flag.NewFlagSet("pulse", flag.ExitOnError)
	port := fs.Int("port", 9090, "server port")
	data := fs.String("data", "", "data directory (default: ~/.pulse)")
	fs.Parse(os.Args[1:])

	srv, err := server.New(*port, *data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pulse: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "pulse on http://localhost:%d\n", *port)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "pulse: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `pulse %s - connect everything. your ai does the rest.

usage:
  pulse [flags]     start the server
  pulse version      print version
  pulse help         show this message

flags:
  -port int          server port (default 9090)
  -data string       data directory (default: ~/.pulse)

examples:
  pulse
  pulse -port 8080
  pulse -data /path/to/data
`, version)
}
