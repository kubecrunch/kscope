package app

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func sanitizeHeaders(headers *map[string]string, duct *map[string]string) error {
	return nil
}

// replaceParams replaces parameters from a string with their relevant values
// if formatted text is "{greet} chandan" and the duct is {greet:hello} then the returned string will be "hello chandan"
func replaceParams(formatted string, duct *map[string]string) string {
	parameters := getParameters(formatted)
	values := make([]string, 0)
	for _, param := range parameters {
		if val, ok := (*duct)[param]; ok {
			values = append(values, fmt.Sprintf("{%s}", param))
			values = append(values, val)
		}
		values = append(values)
	}
	r := strings.NewReplacer(values...)

	return r.Replace(formatted)
}

func getParameters(str string) []string {
	re := regexp.MustCompile(`{([^{}]+)}`)
	result := make([]string, 0)
	matches := re.FindAll([]byte(str), -1)

	for _, val := range matches {
		str := string(val)
		result = append(result, str[1:len(str)-1])
	}

	return result
}

func loadConfig(flowCfg *FlowConfiguration, configLocation string) error {
	configReader := viper.New()

	configReader.SetConfigName("configuration")
	configReader.SetConfigFile(configLocation)

	if err := configReader.ReadInConfig(); err != nil {
		return err
	}

	if err := configReader.Unmarshal(flowCfg); err != nil {
		return err
	}

	return nil

}

func loadBootstrappedSecrets(res *map[string]string, loc string) (err error) {
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

	if err = json.Unmarshal(data, res); err != nil {
		return err
	}

	return nil
}
