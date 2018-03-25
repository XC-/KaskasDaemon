package SSE

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type SSEServer struct {
	connections      map[chan string]bool
	incomingClients  chan chan string
	removableClients chan chan string
	messageQueue     chan string
}

func (s *SSEServer) Start() {
	go func() {
		for {
			select {
			case c := <-s.incomingClients:
				s.connections[c] = true
				log.Println("New connection...")

			case c := <-s.removableClients:
				delete(s.connections, c)
				close(c)
				log.Println("Removed a connection...")

			case msg := <-s.messageQueue:
				for c, _ := range s.connections {
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
	s.incomingClients <- channel
	notify := writer.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		s.removableClients <- channel
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
	server.Start()
	http.Handle(sseEndpoint, server)
	http.ListenAndServe(listenAddress+":"+strconv.Itoa(listenPort), nil)
	return server
}
