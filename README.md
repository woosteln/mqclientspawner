MQCLIENTSPAWNER
===============

A small app to create a batch of dummy mqtt clients that will switch on and 
off and publish random messages to mqtt over a period of time.

Usage
-----

Run the service and it will create N dummy clients.

These clients will publish to a topic of your choice. You can set the average expected
messages over the duration of the lifetime of a client, and when a client is expected 
to peak its message flow.

Clients will connect and disconnect during their lifecycle based on the connectivity
parameter.

Each client will be created with its own generated ClientID property.

Each client is spawned in its own goroutine, which should allow for running a lot of
little clients concurrently.

This app can be used to load-test or explore failure points in a mqtt setup.

Install & run
-------------

You can run standalone or from a docker container.

You configure the service using command line flags or environment variables.

|         Env         |        Flag         |        Type        |       Default        |                                                     Description                                                     |
| ------------------- | ------------------- | ------------------ | -------------------- | ------------------------------------------------------------------------------------------------------------------- |
| MQTT_BROKER         | mqtt_broker         | string             | tcp://mqtt:1883      | MQTT broker address ( note websockets supported if you use ws:// or wss:// protocol)                                |
| MQTT_TOPIC          | mqtt_topic          | string             | client/{{.ClientID}} | Go format string to represent topic to publish to                                                                   |
| MQTT_USER           | mqtt_user           | string             | ""                   | Username for mqtt                                                                                                   |
| MQTT_PASSWORD       | mqtt_password       | string             | ""                   | Password for mqtt                                                                                                   |
| NUM_CLIENTS         | num_clients         | int                | 10                   | Number of clients to create                                                                                         |
| DURATION            | duration            | go duration string | 10m                  | How long to run the app for                                                                                         |
| START_PEAK          | start_peak          | go duration string | 1m                   | When to start peak period                                                                                           |
| END_PEAK            | end_peak            | go duration string | 9m                   | When to end peak period                                                                                             |
| AVG_MESSAGES        | avg_messages        | int                | 80                   | Average number of messages to send per client over the course of duration                                           |
| PEAK_DISTRIBUTION   | peak_distribution   | float              | 0.8                  | Proportion of messages that should be sent during peak                                                              |
| CONNECTION_STRENGTH | connection_strength | float              | 0.9                  | Proportion of duration that clients should spend connected. Clients will connect and disconnect at random intervals |

### Standalone

#### Go get and install

```
go get -u github.com/woosteln/mqclientspawner
```

#### Run

```
mqclientspawner --mqtt_broker=tcp://localhost:1083 --mqtt_topic=device/{{.ClientID}}
```

### Docker

Use the prebuilt docker image.

_cmd line_

```
docker run -d -e MQTT_BROKER=tcp://mqtt:1883 -MQTT_TOPIC=device/{{.ClientID}} woosteln/mqclientspawner:latest
```

_docker compose_

An example stack which gets a mqtt broker and client spawner up and running.

We attach cadvisor so you can check cpu and mem usage of the containers.

```
version: "3"
services:

    mqtt:
        image: emq:latest
        ports:
            - 18083:18083
            - 1883:1883

    mqclientspawner:
        image: woosteln/mqclientspawner:latest
        depends_on:
            - mqtt
        environment:
            MQTT_BROKER: tcp://mqtt:1883
            MQTT_TOPIC: device/{{.ClientID}}
            MQTT_USER: clientspwaner
            NUM_CLIENTS: 500
            DURATION: 10m
            START_PEAK: 1m
            END_PEAK: 9m
            AVG_MESSAGES: 100
            PEAK_DISTRIBUTION: 0.8
            CONNECTION_STRENGTH: 0.8

    cadvisor:
        image: google/cadvisor:latest
        ports:
            - 8080:8080
        volumes:
            - /var/run:/var/run:rw
            - /sys:/sys:ro
            - /var/lib/docker/:/var/lib/docker:ro
```