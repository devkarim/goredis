package main

import (
	"io"
	"log/slog"
	"net"

	"github.com/devkarim/goredis/resp"
)

const DEFAULT_LISTEN_ADDR = ":6379"

type Config struct {
	ListenAddr string
}

type Server struct {
	Config
	ln net.Listener
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = DEFAULT_LISTEN_ADDR
	}
	return &Server{
		Config: cfg,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	slog.Info("Server running at localhost:6379")
	s.ln = ln

	return s.loop()
}

func (s *Server) loop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("Error when accepting client", "error", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := resp.NewReader(conn)
	writer := resp.NewWriter(conn)
	slog.Info("Connection from", "remoteAddr", conn.RemoteAddr().String())
	for {
		message, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				slog.Error("Error while reading from connection", "error", err)
			}
			break
		}
		slog.Info("Received", "message", message)
		if len(message.Array) < 1 {
			writer.Write(resp.Value{Type: resp.RespError, Str: "Command not found"})
			continue
		}
		handler, ok := Handlers[message.Array[0].Str]
		if !ok {
			writer.Write(resp.Value{Type: resp.RespError, Str: "Command not found"})
			continue
		}
		writer.Write(handler(message.Array))
	}
	slog.Info("Disconnected", "remoteAddr", conn.RemoteAddr().String())
}
