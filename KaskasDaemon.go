// +build

package main

import (
	"flag"

	"github.com/XC-/KaskasDaemon/SSE"
)

func main() {
	conf := flag.String("c", "", "Path to the configuration file")
	flag.Parse()
	configuration := configparser.GetConfiguration(*conf)
	var devicesToListen map[string]bool = make(map[string]bool)
	for _, device := range configuration.Devices.Listen {
		devicesToListen[device] = true
	}

	var server *SSE.SSEServer

	if configuration.HTTP.ServeSSE {
		server = SSE.StartHTTP(configuration.HTTP.Listen.Address, configuration.HTTP.Listen.Port, configuration.HTTP.Listen.SSEEndpoint)
	}

	btChannel := make(chan string)
	ruuvireader.StartBT(btChannel, devicesToListen)

	go func() {
		for {
			select {
			case msg := <-btChannel:
				server.MessageQueue <- msg
			}
		}
	}()

	select {}
}
