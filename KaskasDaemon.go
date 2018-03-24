// +build

package main

import (
    "fmt"
    "time"
    "log"
    "net/http"
    "encoding/hex"
    "encoding/binary"
    "github.com/paypal/gatt"
    "github.com/paypal/gatt/examples/option"
)

type Broker struct {
    clients map[chan string]bool
    newClients chan chan string
    oldClients chan chan string
    messages chan string
}

var b *Broker
var devicesToListen map[string]bool

func (b *Broker) Start() {
    go func() {
        for {
            select {
            case s := <-b.newClients:
                b.clients[s] = true
                log.Println("Added new client")

            case s := <-b.oldClients:
                delete(b.clients, s)
                close(s)
                log.Println("Removed client")

            case msg := <-b.messages:
                log.Printf(msg)
                for s, _ := range b.clients {
                        log.Printf("Sending to a client")
                        s <- msg
                }
            }
        }
    }()
}

func (c *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    f, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }
    messageChan := make(chan string)
    fmt.Printf("before new clients", c)
    c.newClients <- messageChan
    log.Printf("After")
    notify := w.(http.CloseNotifier).CloseNotify()
    go func() {
        <-notify
        // Remove this client from the map of attached clients
        // when `EventHandler` exits.
        c.oldClients <- messageChan
        log.Println("HTTP connection just closed.")
    }()

    // Set the headers related to event streaming.
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    for {
        log.Printf("In for loop")
        msg, open := <-messageChan

        if !open {
            // If our messageChan was closed, this means that the client has
            // disconnected.
            break
        }
        fmt.Fprintf(w, "data: %s\n\n", msg)

        f.Flush()
    }
    log.Println("Finished HTTP request at ", r.URL.Path)
}

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
            pressure := float32(binary.BigEndian.Uint16(a.ManufacturerData[6:8])) / 100 + 500
            fmt.Println(humidity, temperature_integer, temperature_decimal, pressure)
            b.messages <- fmt.Sprintf("{" +
                                          "\"deviceId\": \"%s\", " +
                                          "\"timestamp\": %d, " +
                                          "\"rawData\": \"%s\", " +
                                          "\"measurements\": { " +
                                              "\"temperature\": { \"value\": %s%d.%02d, \"unit\": \"celsius\" },  " +
                                              "\"humidity\": {\"value\": %d, \"unit\": \"percent\" }, " +
                                              "\"pressure\": {\"value\": %.2f, \"unit\": \"hPa\" } " +
                                          "} " +
                                      "}", p.ID(), timestamp, hex.EncodeToString(a.ManufacturerData), 
                                          temperature_sign, temperature_integer, temperature_decimal, humidity, pressure)
        }()
    }
}

func startBT() {
    go func() {
        d, err := gatt.NewDevice(option.DefaultClientOptions...)
        if err != nil {
            log.Fatalf("Failed to open device, err: %s\n", err)
            return
        }
        // Register handlers.
        d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))

        d.Init(onStateChanged)
    }()
}

func startHTTP() {
    go func() {
        fmt.Println(b)
        http.Handle("/events/", b)
        fmt.Println("Starting http server")
        go http.ListenAndServe(":27911", nil)
    }()
}

func main() {
    b = &Broker{
        make(map[chan string]bool, 1000),
        make(chan (chan string), 1000),
        make(chan (chan string), 1000),
        make(chan string, 1000000),
    }
    devicesToListen = map[string]bool {
        "C3:BC:E8:BF:6C:AC": true,
        "DE:FD:4A:E0:0A:91": true,
        "DF:61:03:50:8A:60": true,
    }
    fmt.Println(b)
    b.Start()

    startHTTP()
    startBT()

    select {}
}


