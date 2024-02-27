package app

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Params struct {
	MqttBrokerHost string
	MqttBrokerPort int
	CacheTimeout   time.Duration
}

var (
	params   Params
	slaves   = []string{"14", "19", "229", "23", "234", "231", "240", "245"}
	channels = []string{"1", "2", "3", "4"}
	cache    = NewCache()
	mtx      = sync.RWMutex{}
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func sub(client mqtt.Client, slaveID, channel string) {
	topic := fmt.Sprintf("/devices/wb-map12e_%s/controls/Ch %s Total AP energy", slaveID, channel)
	token := client.Subscribe(topic, 1, func(client mqtt.Client, message mqtt.Message) {
		if v64, err := strconv.ParseFloat(string(message.Payload()), 10); err != nil {
			logrus.Errorln(errors.Wrap(err, `mqtt received topic values convert to integer error`))
		} else {
			metric := cache.GetOrCreateMetric(message.Topic())
			if metric.Set(v64) {
				logrus.Infof(`%s: %f`, message.Topic(), v64)
			}
		}
	})
	token.Wait()
	logrus.Infof("Subscribed to topic: %s\n", topic)
}

func clientOptions() *mqtt.ClientOptions {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf(`tcp://%s:%d`, params.MqttBrokerHost, params.MqttBrokerPort))
	options.SetClientID(`tomm-golang-client`)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler
	return options
}

func App(p Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	go CatchSigTerm(ctx, cancel)
	params = p
	client := mqtt.NewClient(clientOptions())
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for _, slaveID := range slaves {
		for _, channel := range channels {
			sub(client, slaveID, channel)
		}
	}
	<-ctx.Done()
	return nil
}
