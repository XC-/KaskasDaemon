package ruuvireader

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/paypal/gatt"
)

// From https://github.com/paypal/gatt/blob/master/examples/option/option_linux.go
var DefBTOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
}

var messageChannel chan string
var devicesToListen map[string]bool

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		d.Scan([]gatt.UUID{}, true)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	if _, ok := devicesToListen[p.ID()]; ok {
		go func() {
			humidity := uint8(a.ManufacturerData[3])
			var temperature_sign string = ""

			if (a.ManufacturerData[4] & 0x80) == 0x80 {
				temperature_sign = "-"
			}

			temperature_integer := int8(a.ManufacturerData[4] & 0x7F)
			temperature_decimal := uint8(a.ManufacturerData[5])
			pressure := float32(binary.BigEndian.Uint16(a.ManufacturerData[6:8]))/100 + 500
			fmt.Println(humidity, temperature_integer, temperature_decimal, pressure)
			messageChannel <- fmt.Sprintf("{"+
				"\"deviceId\": \"%s\", "+
				"\"timestamp\": %d, "+
				"\"rawData\": \"%s\", "+
				"\"measurements\": { "+
				"\"temperature\": { \"value\": %s%d.%02d, \"unit\": \"celsius\" },  "+
				"\"humidity\": {\"value\": %d, \"unit\": \"percent\" }, "+
				"\"pressure\": {\"value\": %.2f, \"unit\": \"hPa\" } "+
				"} "+
				"}", p.ID(), timestamp, hex.EncodeToString(a.ManufacturerData),
				temperature_sign, temperature_integer, temperature_decimal, humidity, pressure)
		}()
	}
}

func StartBT(channel chan string, devices map[string]bool) {
	messageChannel = channel
	devicesToListen = devices
	go func() {
		d, err := gatt.NewDevice(DefBTOptions...)
		if err != nil {
			log.Fatalf("Failed to open device, err: %s\n", err)
			return
		}
		// Register handlers.
		d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))

		d.Init(onStateChanged)
	}()
}
