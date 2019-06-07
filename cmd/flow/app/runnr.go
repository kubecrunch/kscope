package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kubecrunch/kscope/api/v1alpha1"
)

var flowCmd = cobra.Command{
	Use: "flow",
	Run: run,
}

// FlowConfiguration
type FlowConfiguration struct {
	Stages []v1alpha1.KscopeStage `json:"stages"`
}

// PrometheusMetricsHolder holds relevant information about a request response while will be passed to prometheus for monitoring and alerting purposes
type PrometheusMetricsHolder struct {
	StageName          string `json:"stage_name"`
	Url                string `json:"url"`
	Method             string `json:"method"`
	ExpectedStatusCode int    `json:"expected_status_code"`
	ActualStatusCode   int    `json:"actual_status_code"`
	Latency            int64  `json:"latency"`
	AllowedLatency     int    `json:"allowed_latency"`
	ErrorMessage       string `json:"error_message"`
	ResponseText       string `json:"response_text"`
}

// TODO: use logging framework like logrus etc.

// RootCommand will setup and return the root command
func NewFlowCommand() *cobra.Command {
	return &flowCmd
}

func run(cmd *cobra.Command, _ []string) {
	// TODO: should come as a flag
	configFile := "/tmp/config/stages.json"

	cfg := FlowConfiguration{}

	if err := loadConfig(&cfg, configFile); err != nil {
		panic(err)
	}

	duct := make(map[string]string)
	// TODO: bootstrap cnfigs should come as flag
	if err := loadBootstrappedSecrets(&duct, ""); err != nil {
		panic(err)
	}

	// TODO: this should be configurable
	loop(10000*time.Millisecond, cfg, &duct)
}

// loop checks a Linearly independent path every d intervals
func loop(d time.Duration, flowCfg FlowConfiguration, duct *map[string]string) {
	sort.Slice(flowCfg.Stages, func(i, j int) bool {
		return flowCfg.Stages[i].SequenceNumber < flowCfg.Stages[j].SequenceNumber
	})

	var wg sync.WaitGroup

	for _ = range time.Tick(d) {
		for _, stage := range flowCfg.Stages {
			wg.Add(1)
			ctx := context.Background()
			prom, err := handleStage(ctx, &wg, &stage, duct)
			wg.Wait()
			// publish the stage information to prometheus
			fmt.Println(prom) // todo: remove me
			if err != nil {
				// cancel the flow of the LIP
			}
		}
	}

}

// handleStage handles a single kubescope stage which includes visiting a url, capturing and reporting
// useful metrics to prometheus
func handleStage(ctx context.Context, wg *sync.WaitGroup, stage *v1alpha1.KscopeStage,
	duct *map[string]string) (PrometheusMetricsHolder, error) {
	defer wg.Done()

	url := replaceParams(stage.Request.Url, duct)

	prom := PrometheusMetricsHolder{
		StageName: stage.Name,
		Url:       url,
		Method:    stage.Request.Method,
	}

	t := make(chan interface{})
	headers := make(map[string]string)
	for k, v := range stage.Request.Headers {
		s := replaceParams(v, duct)
		headers[k] = s
	}

	go func() {
		defer close(t)

		switch method := stage.Request.Method; method {
		case "GET":
			response, duration := visit("GET", url, &headers, nil)
			prom.ActualStatusCode = response.StatusCode
			prom.ExpectedStatusCode = stage.Response.StatusCode
			prom.Latency = int64(duration / time.Millisecond)

			check(stage, response, duct, &prom)

		case "POST":
			body, err := base64.StdEncoding.DecodeString(stage.Request.Body)
			if err != nil {
				prom.ErrorMessage = "Error!! while decoding request body."
				return
			}
			body = []byte(replaceParams(string(body), duct))

			response, duration := visit("POST", url, &headers, body)
			prom.ActualStatusCode = response.StatusCode
			prom.ExpectedStatusCode = (*stage).Response.StatusCode
			prom.Latency = int64(duration / time.Millisecond)
			check(stage, response, duct, &prom)
		case "PUT":
			// handle PUT call
		case "DELETE":
			// handle delete
		}

	}()

	select {
	case <-t:
		fmt.Println("did it in time")
	case <-time.After(15 * time.Second):
		fmt.Println("timeout, api didn't respond within 15 seconds")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}
	return prom, nil
}

// check checks the response and preserves relevant fields for further stages
func check(stage *v1alpha1.KscopeStage, response *http.Response, duct *map[string]string, prom *PrometheusMetricsHolder) {
	responseBody, err := ioutil.ReadAll(response.Body)
	prom.ResponseText = string(responseBody)
	if err != nil {
		prom.ErrorMessage = "Error!! reading response body."
		return
	}
	if response.StatusCode != stage.Response.StatusCode {
		prom.ErrorMessage = "Error!! status code doesn't match."
		return
	}

	var responseMap interface{}
	json.Unmarshal(responseBody, &responseMap)

	fields := stage.Response.ExpectedFields

	m := responseMap.(map[string]interface{})
	for _, field := range fields {
		chain := strings.Split(field, ".")
		finalKey := chain[len(chain)-1]
		for _, e := range chain[:len(chain)-1] {
			if v, ok := m[e]; ok {
				m = v.(map[string]interface{})

			} else {
				prom.ErrorMessage = fmt.Sprintf("Error!! field %s is not present.", field)
				return
			}

		}
		var v interface{}
		v, ok := m[finalKey]

		if !ok {
			prom.ErrorMessage = fmt.Sprintf("Error!! field %s is not present.", field)
			return
		}
		// check if field is also present in preserved fields
		for _, p := range stage.Response.PreserveFields {
			if p.FieldName == field {
				(*duct)[p.KeyName] = v.(string)
			}
		}

	}

}

// visit visits a url and captures duration etc.
func visit(verb, url string, headers *map[string]string, body []byte) (*http.Response, time.Duration) {
	client := http.Client{}

	req, err := http.NewRequest(verb, url, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	for k, v := range *headers {
		req.Header.Add(k, v)
	}

	start := time.Now()
	response, err := client.Do(req)
	elapsed := time.Since(start)
	return response, elapsed
}

// replaceParams replaces parameters from a string with their relevant values
// if formatted text is "{greet} CK" and the duct is {greet:hello} then the returned string will be "hello CK"
func replaceParams(formatted string, duct *map[string]string) string {
	parameters := getParameters([]byte(formatted))
	values := make([]string, 0)
	for _, param := range parameters {
		if val, ok := (*duct)[param]; ok {
			values = append(values, fmt.Sprintf("{{%s}}", param))
			values = append(values, val)
		}
		values = append(values)
	}
	r := strings.NewReplacer(values...)

	return r.Replace(formatted)
}

// getParameters returns the list of parameters from a formatted string
// eg. for string "Bearer {token}" `getParameters` will return ["token"]
func getParameters(b []byte) []string {
	re := regexp.MustCompile(`{{([^{}]+)}}`)
	result := make([]string, 0)
	matches := re.FindAll([]byte(b), -1)

	for _, val := range matches {
		str := string(val)
		result = append(result, str[2:len(str)-2])
	}

	return result
}

func loadConfig(flowCfg *FlowConfiguration, configLocation string) error {

	if _, err := os.Stat(configLocation); os.IsNotExist(err) {
		return err
	}
	data, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, flowCfg); err != nil {
		return err
	}

	return nil
}

func loadBootstrappedSecrets(duct *map[string]string, loc string) (err error) {
	if loc == "" {
		loc = "/tmp/secrets/bootstrap.json"
	}
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		return err
	}
	data, err := ioutil.ReadFile(loc)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, duct); err != nil {
		return err
	}

	return nil
}
