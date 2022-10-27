package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	StatusCodes []struct {
		StatusCode int `yaml:"code"`
		Repeat     int `yaml:"repeat"`
	} `yaml:"status_codes"`
	Metrics []struct {
		Identifier string `yaml:"identifier"`
		ValueCycle struct {
			InitialValue int `yaml:"initial_value"`
			Trends       []struct {
				Type   string `yaml:"type"`
				Step   int    `yaml:"step"`
				Repeat int    `yaml:"repeat"`
			} `yaml:"trends"`
		} `yaml:"value_cycle"`
	} `yaml:"metrics"`

	metricCounters []struct {
		ActualValue   int
		ActualCounter int
		ActualTrend   int
	}
	actualStatusCodeIndex   int
	actualStatusCodeCounter int
	requestCounter          int
}

func (c *Config) getResponseBody() (body string) {
	sb := strings.Builder{}
	for index, metric := range c.Metrics {
		sb.WriteString(metric.Identifier)
		sb.WriteString(" ")
		sb.WriteString(c.getActualValueFromMetric(index))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (c *Config) getActualValueFromMetric(metricIndex int) string {
	return strconv.Itoa(c.metricCounters[metricIndex].ActualValue)
}

func (c *Config) getActualReturnCode() int {
	return c.StatusCodes[c.actualStatusCodeIndex].StatusCode
}

func (c *Config) stepStatusCode() {
	if c.actualStatusCodeCounter < c.StatusCodes[c.actualStatusCodeIndex].Repeat-1 {
		c.actualStatusCodeCounter++
	} else if c.actualStatusCodeIndex < len(c.StatusCodes)-1 {
		c.actualStatusCodeCounter = 0
		c.actualStatusCodeIndex++
	} else {
		c.actualStatusCodeCounter = 0
		c.actualStatusCodeIndex = 0
	}
}

func (c *Config) nextValueOfMetric(index int) {
	if c.Metrics[index].ValueCycle.Trends[c.metricCounters[index].ActualTrend].Type == "increment" {
		c.metricCounters[index].ActualValue += c.Metrics[index].ValueCycle.Trends[c.metricCounters[index].ActualTrend].Step
	} else {
		c.metricCounters[index].ActualValue -= c.Metrics[index].ValueCycle.Trends[c.metricCounters[index].ActualTrend].Step
	}
}

func (c *Config) stepMetric(index int) {
	if c.metricCounters[index].ActualCounter < c.Metrics[index].ValueCycle.Trends[c.metricCounters[index].ActualTrend].Repeat {
		c.metricCounters[index].ActualCounter++
		c.nextValueOfMetric(index)
	} else if c.metricCounters[index].ActualTrend < len(c.Metrics[index].ValueCycle.Trends)-1 {
		c.metricCounters[index].ActualTrend++
		c.metricCounters[index].ActualCounter = 0
		c.nextValueOfMetric(index)
	} else {
		c.metricCounters[index].ActualCounter = 0
		c.metricCounters[index].ActualTrend = 0
		c.nextValueOfMetric(index)
	}
}

// increment all counters
func (c *Config) prepareForNextRequest() {
	c.stepStatusCode()
	for index := range c.Metrics {
		c.stepMetric(index)
	}
}

func (c *Config) init() {
	file, err := os.Open(configFilePath)
	defer file.Close()
	if err != nil {
		log.Println(err.Error())
	}

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		log.Println(err.Error())
	}

	c.requestCounter = 0
	c.actualStatusCodeCounter = 0
	c.actualStatusCodeIndex = 0
	c.metricCounters = make([]struct {
		ActualValue   int
		ActualCounter int
		ActualTrend   int
	}, len(c.Metrics))
	for index, metric := range c.Metrics {
		c.metricCounters[index].ActualValue = metric.ValueCycle.InitialValue
		c.metricCounters[index].ActualCounter = 0
		c.metricCounters[index].ActualTrend = 0
	}
}

var conf Config

var configFilePath string

// no thread safe (1 request at a time or crash)
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(conf.getActualReturnCode())
	fmt.Fprint(w, conf.getResponseBody())
	conf.prepareForNextRequest()
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Config file not specified, using default /app/config.yml")
		configFilePath = "/app/config.yml"
	} else {
		configFilePath = os.Args[1]
	}
	conf.init()
	fmt.Printf("%+v\n", conf)

	http.HandleFunc("/metrics", metricsHandler)

	http.ListenAndServe(":5000", nil)
}
