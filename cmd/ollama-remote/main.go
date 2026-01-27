package main

import (
	"os"

	"cli_ollama_server/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
