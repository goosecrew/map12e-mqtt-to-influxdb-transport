package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
	"tomm-map12e/app"
)

func main() {
	params := app.Params{}
	cmd := cobra.Command{
		Use: `mqtt-to-influxdb-transport`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := app.App(params); err != nil {
				logrus.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringVar(&params.MqttBrokerHost, `mqtt-broker-host`, `127.0.0.1`, `адрес mqtt брокера`)
	cmd.PersistentFlags().IntVar(&params.MqttBrokerPort, `mqtt-broker-port`, 1883, `порт mqtt брокера`)
	cmd.PersistentFlags().DurationVar(&params.CacheTimeout, `cache-timeout`, time.Second*5, `как долго хранить кэш значений по каждому элементу`)
	cmd.Execute()
}
