package main

import (
	"log"

	"github.com/sebrandon1/openshift-preflight/cmd/preflight/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
