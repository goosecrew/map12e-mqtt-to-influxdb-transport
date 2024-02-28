package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb1 "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Params struct {
	MqttBrokerHost      string
	MqttBrokerPort      int
	CacheTimeout        time.Duration
	InfluxDBHost        string
	InfluxDBPort        int
	InfluxDBToken       string
	InfluxDBOrg         string
	InfluxDBBucket      string
	InfluxDBMeasurement string
}

var (
	params      Params
	slaves      = []string{"14", "19", "229", "23", "234", "231", "240", "245"}
	channels    = []string{"1", "2", "3", "4"}
	cache       = NewCache()
	mtx         = sync.RWMutex{}
	db          influxdb2.Client
	ctx, cancel = context.WithCancel(context.Background())
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

func sub(client mqtt.Client, slaveID, channel, suffix string) {
	topic := fmt.Sprintf("/devices/wb-map12e_%s/controls/Ch %s %s", slaveID, channel, suffix)
	moscowLocation, err := time.LoadLocation(`Europe/Moscow`)
	if err != nil {
		logrus.Fatalln(err)
	}
	token := client.Subscribe(topic, 1, func(client mqtt.Client, message mqtt.Message) {
		if v64, err := strconv.ParseFloat(string(message.Payload()), 10); err != nil {
			logrus.Errorln(errors.Wrap(err, `mqtt received topic values convert to integer error`))
		} else {
			if v64 > 0 {
				metric := cache.GetOrCreateMetric(message.Topic())
				if metric.Set(v64) {
					//logrus.Infof(`%s: %f`, message.Topic(), v64)
					writeAPI := db.WriteAPIBlocking(params.InfluxDBOrg, params.InfluxDBBucket)
					now := time.Now()
					p := influxdb2.NewPointWithMeasurement(params.InfluxDBMeasurement).
						AddTag("topic", message.Topic()).
						AddField("value", v64).
						AddField("hour", now.In(moscowLocation).Hour()).
						SetTime(now)
					logrus.Infof(`%s write topic=%s, value=%f, hour=%d`, now, message.Topic(), v64, now.In(moscowLocation).Hour())
					if err := writeAPI.WritePoint(ctx, p); err != nil {
						logrus.Errorln(err)
					}
				}
			}
		}
	})
	token.Wait()
	logrus.Infof("Subscribed to topic: %s\n", topic)
}

func mqttClientOptions() *mqtt.ClientOptions {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf(`tcp://%s:%d`, params.MqttBrokerHost, params.MqttBrokerPort))
	options.SetClientID(`tomm-golang-client`)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler
	return options
}

func Daemon(p Params) error {
	go CatchSigTerm(ctx, cancel)
	params = p
	mqttClient := mqtt.NewClient(mqttClientOptions())
	db = influxdb2.NewClient(fmt.Sprintf(`http://%s:%d`, params.InfluxDBHost, params.InfluxDBPort), params.InfluxDBToken)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for _, slaveID := range slaves {
		for _, channel := range channels {
			sub(mqttClient, slaveID, channel, `Total AP energy`)
			sub(mqttClient, slaveID, channel, `Total P`)
		}
	}
	<-ctx.Done()
	return nil
}

type SandboxParams struct {
	V1Addr    string
	V2Addr    string
	Token     string
	StartTime string
	StopTime  string
}

func Sandbox(p SandboxParams) {
	v1config := influxdb1.HTTPConfig{
		Addr: p.V1Addr,
	}
	v1client, err := influxdb1.NewHTTPClient(v1config)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer v1client.Close()

	v2client := influxdb2.NewClient(p.V2Addr, p.Token)
	moscowLocation, err := time.LoadLocation(`Europe/Moscow`)
	if err != nil {
		logrus.Fatalln(err)
	}
	startTime, err := time.Parse(time.RFC3339, p.StartTime)
	stopTime, err := time.Parse(time.RFC3339, p.StopTime)
	currentTime := startTime
	if err != nil {
		logrus.Fatalln(err)
	}
	for {
		nextTime := currentTime.Add(time.Minute * 1)
		if stopTime.Unix() > time.Now().Unix() {
			break
		}
		sql := fmt.Sprintf("SELECT last(value) FROM mqtt_consumer WHERE time >= '%s' AND time <= '%s' GROUP BY topic ORDER BY time\n", currentTime.Format(time.RFC3339), nextTime.Format(time.RFC3339))
		q := influxdb1.Query{
			Command:  sql,
			Database: `telegraf`,
		}
		response, err := v1client.Query(q)
		if err != nil {
			logrus.Fatalln(err)
		}
		for _, item := range response.Results {
			for _, serie := range item.Series {
				topic, ok := serie.Tags[`topic`]
				if !ok {
					logrus.Fatalln(`invalid topic`)
				}
				t, err := time.Parse(time.RFC3339, serie.Values[0][0].(string))
				if err != nil {
					logrus.Fatalln(err)
				}
				value, ok := serie.Values[0][1].(json.Number)
				if !ok {
					logrus.Fatalln(`invalid json.Number`, value)
				}
				v64, err := value.Float64()
				if err != nil {
					logrus.Fatalln(`invalid value`, value)
				}
				writeAPI := v2client.WriteAPIBlocking(`mqtt`, `wirenboard`)
				p := influxdb2.NewPointWithMeasurement(`map12e`).
					AddTag("topic", topic).
					AddField("value", v64).
					AddField("hour", t.In(moscowLocation).Hour()).
					SetTime(t)
				if err := writeAPI.WritePoint(ctx, p); err != nil {
					logrus.Fatalln(errors.Wrapf(err, `FAILED write topic=%s, value=%f, hour=%d`, topic, v64, t.In(moscowLocation).Hour()))
				} else {
					logrus.Infoln(fmt.Sprintf(`%s SUCCESS write topic=%s, value=%f, hour=%d`, t, topic, v64, t.In(moscowLocation).Hour()))
				}
			}
		}
		currentTime = nextTime

	}
	return

}
