package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/namsral/flag"
	"github.com/woosteln/mqclientspawner/dummyclient"
)

var (
	mqttBroker           = flag.String("mqtt_broker", "tcp://mqtt:1883", "MQTT broker address")
	mqttTopic            = flag.String("mqtt_topic", "/client/{{.ClientID}}", "Go template string for topic. Default /client/{{.ClientID}}")
	mqttUser             = flag.String("mqtt_user", "", "Username for mqtt")
	mqttPassword         = flag.String("mqtt_password", "", "Password for mqtt")
	numClients           = flag.Int("num_clients", 10, "Number of dummy clients to create")
	duration             = flag.String("duration", "24h", "How long to run for ( duration string. Default 24h")
	startOfPeak          = flag.String("start_peak", "8h", "Duration after which to start peak flow. Default 8h")
	endOfPeak            = flag.String("end_peak", "22h", "Duration after which to end peak flow. Default 22h")
	avgMessagesPerPeriod = flag.Int("avg_messages", 80, "Average messages to send during duration")
	peakDistribution     = flag.Float64("peak_distribution", 0.8, "Proportion of messages to send during peak")
	connectionStrength   = flag.Float64("connection_strength", 0.9, "Connection strength per client. They will randomly switch off and on the connection according to this rate")
)

func main() {

	flag.Parse()

	totalMessages := make(chan dummyclient.DummyClientResult, *numClients)
	var wg sync.WaitGroup
	for i := 0; i < *numClients; i++ {
		clientUUID, _ := uuid.NewUUID()
		dummyClient := dummyclient.NewDummyClient(*mqttBroker, clientUUID.String(), *mqttTopic, *mqttUser, *mqttPassword, *duration, *startOfPeak, *endOfPeak, *avgMessagesPerPeriod, *peakDistribution, *connectionStrength)
		wg.Add(1)
		go func(cnts chan dummyclient.DummyClientResult, d dummyclient.DummyClient, wg *sync.WaitGroup) {
			defer wg.Done()
			cnts <- d.DoLifecycle()
		}(totalMessages, dummyClient, &wg)
	}
	wg.Wait()
	fmt.Println("Done")
	for r := range totalMessages {
		b, _ := json.Marshal(r)
		fmt.Printf("%s\n", string(b))
	}

}
