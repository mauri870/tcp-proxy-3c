package main

import (
	"bufio"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	logLevel         = getEnv("TCP_PROXY_LOG_LEVEL", "INFO")
	eventsServerAddr = getEnv("EVENTS_SERVER_ADDR", "")
	addr             = getEnv("TCP_PROXY_ADDR", "")

	pingInterval = 5 * time.Second
	pingMessage  = []byte("PING\r\n")
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
	log.Debugf("Handling conn from %s\n", c.RemoteAddr().String())
	for {
		token, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			// If the client does not provide a token then bail out
			log.Error(err)
			break
		}
		token = strings.TrimSpace(string(token))

		ws, err := setupWsStream(token)
		if err != nil {
			log.Error(err)
			break
		}

		streaming := ws.Start()
		defer ws.Stop()

		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for {
			select {
			case data := <-streaming:
				data = append(data, []byte("\r\n")...)
				c.Write(data)
			case <-ticker.C:
				_, err := c.Write(pingMessage)
				if err != nil {
					log.Error(err)
					break
				}
			}
		}
	}
	c.Close()
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