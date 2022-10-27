# prometheus-mock-service

So you want to learn PromQL and you don't have fake data for creating queries? Or you want to understand how Prometheus alerts work but you find yourself shutting down services in order to test the alarm triggers?

Set up this service as a mock for being monitored by Prometheus. With it you will be able to configure metrics, their values, and the evolution of those values over time. Therefore you will be able to test Prometheus triggers and queries on a service with a previously set behaviour.

The configuration of the mock is divided in two dimensions, the first is metrics and the second is http return codes. 

Each time Prometheus requests the "/metrics" endpoint, every metric configured calculates its value and put it on the response. The same happens with the http return codes.

In order to configure the mock, you should use a yaml config like this example:




```{yaml}
metrics:
  -
    identifier: "requests{code=\"200\", method=\"GET\"}"
    value_cycle:
      initial_value: 10
      trends:
        -
          repeat: 10
          step: 1
          type: "increment"
        -
          repeat: 3
          step: 3
          type: "decrement"
        -
          repeat: 7
          step: 5
          type: "increment"
        -
          repeat: 20
          step: 0
          type: "decrement"
  -
    identifier: "requests{code=\"200\", method=\"POST\"}"
    value_cycle:
      initial_value: 10
      trends:
        -
          repeat: 10
          step: 1
          type: "increment"
        -
          repeat: 3
          step: 3
          type: "decrement"
        -
          repeat: 7
          step: 5
          type: "increment"
        -
          repeat: 20
          step: 0
          type: "decrement"



status_codes:
  -
    code: 200
    repeat: 100
  -
    code: 500
    repeat: 2
```
