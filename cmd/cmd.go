package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
	"tomm-map12e/app"
)

func main() {
	cmd := cobra.Command{
		Use: `mqtt-to-influxdb-transport`,
	}
	cmd.AddCommand(appCmd())
	cmd.AddCommand(sandboxCmd())
	cmd.Execute()
}

func appCmd() *cobra.Command {
	params := app.Params{}
	cmd := cobra.Command{
		Use: `daemon`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := app.Daemon(params); err != nil {
				logrus.Fatal(err)
			}
		},
	}
	cmd.PersistentFlags().StringVar(&params.MqttBrokerHost, `mqtt-broker-host`, `127.0.0.1`, `адрес mqtt брокера`)
	cmd.PersistentFlags().IntVar(&params.MqttBrokerPort, `mqtt-broker-port`, 1883, `порт mqtt брокера`)
	cmd.PersistentFlags().StringVar(&params.InfluxDBHost, `influxdb-host`, `127.0.0.1`, `адрес influxdb`)
	cmd.PersistentFlags().IntVar(&params.InfluxDBPort, `influxdb-port`, 8086, `порт influxdb`)
	cmd.PersistentFlags().StringVar(&params.InfluxDBToken, `influxdb-token`, `Kp12p4BdYDbE1--wyTfzbiWjQXB62YCAdUwo_tkBBi6Itm59z9Rz751akeEybXGnBCByVchMaj05Kj0mmuMjpA==`, `токен influxdb`)
	cmd.PersistentFlags().StringVar(&params.InfluxDBOrg, `influxdb-org`, `mqtt`, `токен influxdb`)
	cmd.PersistentFlags().StringVar(&params.InfluxDBBucket, `influxdb-bucket`, `wirenboard`, `influxdb bucket`)
	cmd.PersistentFlags().StringVar(&params.InfluxDBMeasurement, `influxdb-measurement`, `map12e`, `influxdb measurement`)
	cmd.PersistentFlags().DurationVar(&params.CacheTimeout, `cache-timeout`, time.Second*5, `как долго хранить кэш значений по каждому элементу`)
	return &cmd

}

func sandboxCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: `sandbox`,
		Run: func(cmd *cobra.Command, args []string) {
			app.Sandbox()
		},
	}
	return &cmd
}
