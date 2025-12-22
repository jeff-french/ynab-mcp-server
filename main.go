package main

import (
	"log"

	"github.com/jeff-french/ynab-mcp-server/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
