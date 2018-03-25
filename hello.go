package main

import (
	"flag"

	"github.com/XC-/KaskasDaemon/ConfigParser"
)

func main() {
	conf := flag.String("configuration", "", "Path to the configuration file.")
	flag.Parse()
	ConfigParser.GetConfiguration(conf)
}
