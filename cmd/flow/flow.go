package main

import (
	"github.com/kubecrunch/kscope/cmd/flow/app"
	"log"
)

func main() {
	if err := app.NewFlowCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
