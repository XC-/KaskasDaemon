package SSE

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type SSEServer struct {
	Connections      map[chan string]bool
	IncomingClients  chan chan string
	RemovableClients chan chan string
	MessageQueue     chan string
}

func (s *SSEServer) Start() {
	log.Println("Starting to serve SSE clients...")
	go func() {
		for {
			select {
			case c := <-s.IncomingClients:
				s.Connections[c] = true
				log.Println("New connection...")

			case c := <-s.RemovableClients:
				delete(s.Connections, c)
				close(c)
				log.Println("Removed a connection...")

			case msg := <-s.MessageQueue:
				for c, _ := range s.Connections {
					log.Println("Sending to client...")
					c <- msg
				}
			}
		}
	}()
}

func (s *SSEServer) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	f, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "SSE Unsupported", http.StatusInternalServerError)
		return
	}
	channel := make(chan string)
	s.IncomingClients <- channel
	notify := writer.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		s.RemovableClients <- channel
		log.Println("HTTP Connection closed")
	}()

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	for {
		msg, open := <-channel
		if !open {
			break
		}
		fmt.Fprintf(writer, "data: %s\n\n", msg)
		f.Flush()
	}
}

func StartHTTP(listenAddress string, listenPort int, sseEndpoint string) *SSEServer {
	server := &SSEServer{
		make(map[chan string]bool, 1000),
		make(chan (chan string), 1000),
		make(chan (chan string), 1000),
		make(chan string, 10000000),
	}
	go func() {
		server.Start()
		http.Handle(sseEndpoint, server)
		http.ListenAndServe(listenAddress+":"+strconv.Itoa(listenPort), nil)
	}()
	return server
}
