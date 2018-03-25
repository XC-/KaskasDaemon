package ConfigParser

import (
	"fmt"
	"io/ioutil"
	"os"
)

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

func parseConfiguration(fileContent []byte) {
	fmt.Printf("%s", string(fileContent))
}

func GetConfiguration(configurationFilePath string) {
	f, e := readConfiguration(configurationFilePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	parseConfiguration(f)
}
