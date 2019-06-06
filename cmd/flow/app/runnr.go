package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"sort"
	"sync"
	"time"

	"github.com/kubecrunch/kscope/api/v1alpha1"
)

var flowCmd = cobra.Command{
	Use: "flow",
	Run: run,
}

// TODO: use logging framework like logrus etc.

// RootCommand will setup and return the root command
func NewFlowCommand() *cobra.Command {
	return &flowCmd
}

func run(cmd *cobra.Command, _ []string) {
	configFile := "/tmp/config/stages.json"

	cfg := FlowConfiguration{}

	if err := loadConfig(&cfg, configFile); err != nil {
		panic(err)
	}

	duct := make(map[string]string)
	if err := loadBootstrappedSecrets(&duct, ""); err != nil {
		panic(err)
	}

	// TODO: this should be configurable
	loop(2000*time.Millisecond, cfg, &duct)

}

func loop(d time.Duration, flowCfg FlowConfiguration, duct *map[string]string) {
	sort.Slice(flowCfg.Stages, func(i, j int) bool {
		return flowCfg.Stages[i].SequenceNumber < flowCfg.Stages[j].SequenceNumber
	})

	var wg sync.WaitGroup

	for _ = range time.Tick(d) {
		for _, stage := range flowCfg.Stages {
			wg.Add(1)
			ctx := context.Background()
			//response := make(map[string]interface{})
			handleStage(ctx, &wg, &stage, duct)
			wg.Wait()
		}
	}

}

func handleStage(ctx context.Context, wg *sync.WaitGroup, stage *v1alpha1.KscopeStage,
	duct *map[string]string) error {
	defer wg.Done()

	var start time.Time
	t := make(chan interface{})
	headers := make(map[string]string)
	for k, v := range stage.Request.Headers {
		headers[k] = replaceParams(v, duct)
	}

	//url := replaceParams(stage.Request.Url, duct)

	//res := http.Response{}

	go func() {
		defer close(t)

		switch method := stage.Request.Method; method {
		case "GET":
			start = time.Now()
			// handle get call
		case "POST":
			start = time.Now()
			// handle POST call
		case "PUT":
			start = time.Now()
			// handle PUT call
		case "DELETE":
			start = time.Now()
			// handle delete
		}

	}()

	select {
	case <-t:
		_ = time.Since(start) // elapsed duration
		fmt.Println("did it in time")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}

	return nil

}
