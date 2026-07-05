package main

import (
	"os"

	"github.com/bizshuk/yfinance-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}