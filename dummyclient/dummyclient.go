package dummyclient

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"text/template"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var dummyMessage = `{"clientID": "%s","messageID": %d,"message":"Hello","ts":%d}`

type DummyClientResult struct {
	MessagesSent   int    `json:"messagesSent"`
	LastMessageID  int    `json:"lastMessageId"`
	ClientID       string `json:"clientId"`
	MessagesUnsent int    `json:"messagesUnsent"`
}

type DummyClient struct {
	MQTT               MQTT.Client
	MQTTOpts           MQTT.ClientOptions
	clientID           string
	duration           time.Duration
	peakStart          time.Duration
	peakEnd            time.Duration
	avgMessages        int
	distribution       float64
	topic              string
	connectionStrength float64
}

func NewDummyClient(broker string, clientID string, topic string, userName string, password string, duration string, peakStart string, peakEnd string, avgMessages int, distribution float64, connectionStrength float64) DummyClient {

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetUsername(userName)
	opts.SetPassword(password)
	opts.SetClientID(clientID)

	lDuration, _ := time.ParseDuration(duration)
	lPeakStart, _ := time.ParseDuration(peakStart)
	lPeakEnd, _ := time.ParseDuration(peakEnd)

	tmpl, err := template.New("test").Parse(topic)
	if err != nil {
		panic(err)
	}
	var data bytes.Buffer
	if err := tmpl.Execute(&data, struct {
		ClientID string
		Username string
		Password string
	}{clientID, userName, password}); err != nil {
		panic(err)
	}
	topic = data.String()

	return DummyClient{
		MQTTOpts:           *opts,
		clientID:           clientID,
		duration:           lDuration,
		peakStart:          lPeakStart,
		peakEnd:            lPeakEnd,
		avgMessages:        avgMessages,
		distribution:       distribution,
		topic:              topic,
		connectionStrength: connectionStrength,
	}
}

func (d *DummyClient) DoLifecycle() DummyClientResult {

	var messages []string

	messagesPeak := int64(float64(d.avgMessages) * d.distribution)
	messagesOffPeak := int64(d.avgMessages) - messagesPeak
	peakDuration := d.peakEnd - d.peakStart
	offPeakDuration := d.duration - peakDuration
	peakRateNs := peakDuration.Nanoseconds() / messagesPeak
	offPeakRateNs := offPeakDuration.Nanoseconds() / messagesOffPeak

	messageID := 0
	messagesSent := 0
	start := time.Now()

	for {

		now := time.Now()
		elapsed := now.Sub(start)

		if elapsed > d.duration {
			break
		}

		inPeak := elapsed > d.peakStart && elapsed < d.peakEnd

		// Add message to queues
		messageID++
		nowMs := now.UnixNano() / 1000000
		messages = append(messages, fmt.Sprintf(dummyMessage, d.clientID, messageID, nowMs))

		// If disconnected, calculate odds of reconnect
		if !d.isConnected() && d.shouldConnectionBeUp() {
			d.connect()
		} else if d.isConnected() && !d.shouldConnectionBeUp() {
			d.disconnect()
		}

		// If connected, send messages
		if d.isConnected() {
			before := len(messages)
			messages = d.sendMessages(messages)
			after := len(messages)
			messagesSent += before - after
		}

		// Calculate next message time
		var nextMessageDelayNs time.Duration
		if inPeak {
			nextMessageDelayNs = time.Duration(getNextNsTime(peakRateNs))
		} else {
			nextMessageDelayNs = time.Duration(getNextNsTime(offPeakRateNs))
		}

		time.Sleep(nextMessageDelayNs)

	}

	if d.MQTT.IsConnected() {
		d.disconnect()
	}

	return DummyClientResult{
		ClientID:       d.clientID,
		MessagesSent:   messagesSent,
		LastMessageID:  messageID,
		MessagesUnsent: len(messages),
	}

}

func (d *DummyClient) isConnected() bool {
	return d.MQTT != nil && d.MQTT.IsConnected()
}

func (d *DummyClient) connect() {
	cli := MQTT.NewClient(&d.MQTTOpts)
	d.MQTT = cli
	cli.Connect()
}

func (d *DummyClient) disconnect() {
	if d.MQTT != nil {
		d.MQTT.Disconnect(240)
		d.MQTT = nil
	}
}

func (d *DummyClient) shouldConnectionBeUp() bool {
	r := rand.Float64()
	return r < d.connectionStrength
}

func getNextNsTime(avg int64) int64 {
	rate := 1.0 / float64(avg)
	return int64(-math.Log(1-rand.Float64()) / rate)
}

func (d *DummyClient) sendMessages(messages []string) []string {
	var failed []string
	for _, msg := range messages {
		if token := d.MQTT.Publish(d.topic, byte(1), false, msg); token.Error() != nil {
			fmt.Printf("Error on client %s: %s\n", d.clientID, token.Error().Error())
			failed = append(failed, msg)
		}

	}
	fmt.Printf("Sent %d messages\n", len(messages)-len(failed))
	return failed
}
