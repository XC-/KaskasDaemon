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
		configurationFiles := []string{"./configuration.conf", "/etc/KaskasDaemon.conf"}
		for _, c := range configurationFiles {
			file, err = ioutil.ReadFile(c)
			if err == nil {
				break
			}
		}
	} else {

	}
	return file, err
}

func parseConfiguration(fileContent []byte) {
	fmt.Printf("%s", string(fileContent))
}

func getConfiguration() {
	f, e := readConfiguration("")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	parseConfiguration(f)
}
