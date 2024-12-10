build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o ./bin/map12e-mqtt-to-influxdb-transport ./cmd
build-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ./bin/map12e-mqtt-to-influxdb-transport ./cmd
