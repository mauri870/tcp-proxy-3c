package main

import (
	"bufio"
	"bytes"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	logLevel         = getEnv("TC_LOG_LEVEL", "INFO")
	eventsServerAddr = getEnv("TC_EVENTS_SERVER_ADDR", "events.3c.fluxoti.com")
	addr             = getEnv("TC_TCP_PROXY_ADDR", ":9090")

	pingInterval = 5 * time.Second
	pingMessage  = []byte("PING\n")
)

func init() {
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
}

func main() {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("TCP server started on address: %s", addr)
	log.Infof("3C Events server address: %s", eventsServerAddr)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	log.Debugf("Handling conn from %s\n", c.RemoteAddr().String())

	token, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		// If the client does not provide a token then bail out
		log.Error(err)
		return
	}
	token = strings.TrimSpace(string(token))

	ws, err := setupWsStream(token)
	if err != nil {
		log.Error(err)
		return
	}

	streaming := ws.Start()
	defer ws.Stop()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	var buf bytes.Buffer
	for {
		select {
		case data := <-streaming:
			buf.Reset()
			buf.Write(data)
			buf.WriteString("\n")
			c.Write(buf.Bytes())
		case <-ticker.C:
			_, err := c.Write(pingMessage)
			if err != nil {
				log.Error(err)
				break
			}
		}
	}
}

func setupWsStream(token string) (EventStream, error) {
	query := url.Values{}
	query.Add("token", token)
	u := url.URL{
		Scheme:   "wss",
		Host:     eventsServerAddr,
		Path:     "/ws/company",
		RawQuery: query.Encode(),
	}
	log.Debug("Ws: ", u.String())

	st, err := NewWebsocketEventStream(u)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// getEnv gets an environment variable or a default
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
