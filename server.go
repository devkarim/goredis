package main

import (
	"io"
	"log/slog"
	"net"
	"strings"

	"github.com/devkarim/goredis/resp"
	"github.com/devkarim/goredis/command"
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
		if message.Type != resp.RespArray {
			writer.Write(resp.Value{Type: resp.RespError, Str: "Invalid request, expected array"})
			continue
		}
		if len(message.Array) <= 0 {
			writer.Write(resp.Value{Type: resp.RespError, Str: "Invalid request, expected array length > 0"})
			continue
		}

		cmd := message.Array[0].Str
		cmdUpper := strings.ToUpper(cmd)
		args := message.Array[1:]

		handler, ok := command.Handlers[cmdUpper]
		if !ok {
			writer.Write(resp.Value{Type: resp.RespError, Str: "ERR unknown command '" + cmd + "'"})
			continue
		}
		writer.Write(handler(args))
	}
	slog.Info("Disconnected", "remoteAddr", conn.RemoteAddr().String())
}
