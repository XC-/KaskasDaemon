package configparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type HTTPListen struct {
	Address     string `json:"address"`
	Port        int    `json:"port"`
	SSEEndpoint string `json:"sse-endpoint"`
}

type HTTP struct {
	Listen   HTTPListen `json:"listen"`
	ServeSSE bool       `json:"serve-sse"`
}

type Devices struct {
	Listen []string `json:"listen"`
}

type Configuration struct {
	HTTP    HTTP    `json:"http"`
	Devices Devices `json:"devices"`
}

var configuration Configuration = Configuration{
	HTTP: HTTP{
		ServeSSE: true,
		Listen: HTTPListen{
			Address:     "0.0.0.0",
			Port:        27911,
			SSEEndpoint: "/events/",
		},
	},
	Devices: Devices{
		Listen: []string{},
	},
}

func readConfiguration(configurationFilePath string) ([]byte, error) {
	var file []byte
	var err error

	if configurationFilePath == "" {
		configurationFiles := []string{"./KaskasDaemon.conf", "/etc/KaskasDaemon.conf"}
		for _, c := range configurationFiles {
			file, err = ioutil.ReadFile(c)
			if err == nil {
				break
			}
		}
	} else {
		file, err = ioutil.ReadFile(configurationFilePath)
	}
	return file, err
}

func parseConfiguration(fileContent []byte) Configuration {
	json.Unmarshal(fileContent, &configuration)
	finalConfigurationJSON, _ := json.Marshal(configuration)
	fmt.Println(string(finalConfigurationJSON))
	return configuration
}

func GetConfiguration(configurationFilePath string) Configuration {
	f, e := readConfiguration(configurationFilePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	return parseConfiguration(f)
}
