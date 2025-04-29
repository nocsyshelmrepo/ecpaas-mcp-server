package main

import (
	"fmt"
	"os"

	"kubesphere.io/ks-mcp-server/cmd/app"
)

func main() {
	if err := app.NewRootCommand().Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
