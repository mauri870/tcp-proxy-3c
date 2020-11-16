package main

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

type EventStream interface {
	Start() chan []byte
	Stop()
}

type WebsocketEventStream struct {
	URL     url.URL
	conn    *websocket.Conn
	stream  chan []byte
	closing chan struct{}
}

func NewWebsocketEventStream(u url.URL) (*WebsocketEventStream, error) {
	s := &WebsocketEventStream{stream: make(chan []byte), closing: make(chan struct{}, 1), URL: u}
	if err := s.connect(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *WebsocketEventStream) connect() error {
	c, _, err := websocket.DefaultDialer.Dial(s.URL.String(), nil)
	if err != nil {
		return err
	}

	s.conn = c
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(15 * time.Second)); return nil })

	return nil
}

func (s *WebsocketEventStream) disconnect() error {
	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}

func (s *WebsocketEventStream) eventLoop() error {
	for {
		select {
		case <-s.closing:
			close(s.closing)
			close(s.stream)
		default:
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				return err
			}
			s.stream <- message
		}
	}
}

func (s *WebsocketEventStream) Start() chan []byte {
	go func() {
		for {
			err := s.eventLoop()
			if err != nil {
				log.Error(err)

				time.Sleep(2 * time.Second)
				if err := s.disconnect(); err != nil {
					log.Error(err)
				}

				if err := s.connect(); err != nil {
					log.Error(err)
				}
			}

		}

	}()
	return s.stream
}

func (s *WebsocketEventStream) Stop() {
	s.closing <- struct{}{}
	s.conn.Close()
}
