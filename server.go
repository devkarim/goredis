package main

import (
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/devkarim/goredis/commands"
	"github.com/devkarim/goredis/core"
	"github.com/devkarim/goredis/eviction"
	"github.com/devkarim/goredis/resp"
	"github.com/devkarim/goredis/storage"
)

const SYNC_TIME = time.Second * 1

type Server struct {
	core.Config
	ln  net.Listener
	aof *storage.Aof
}

func NewServer(cfg core.Config) *Server {
	return &Server{
		Config: cfg,
	}
}

func (s *Server) Start() error {
	storage.Setup(eviction.NewPolicy(s.Policy), s.MaxMemory)

	aof, err := storage.NewAof(s.AofPath)
	if err != nil {
		slog.Error("Couldn't read aof", "error", err)
		return err
	}
	defer aof.Close()

	aof.Read(func(val resp.Value) {
		cmd := strings.ToUpper(val.Array[0].Str)
		command, ok := commands.Registry[cmd]
		if ok {
			slog.Info("Executing from AOF", "command", val)
			args := val.Array[1:]
			command.Handler(args)
		}
	})

	// Sync AOF every some specific duration to prevent OS deciding when to flush the file to disk
	go func() {
		for {
			aof.Sync()
			time.Sleep(SYNC_TIME)
		}
	}()

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	slog.Info("Server running at", "listenAddr", s.ListenAddr)
	s.ln = ln
	s.aof = aof

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

		command, ok := commands.Registry[cmdUpper]
		if !ok {
			writer.Write(resp.Value{Type: resp.RespError, Str: "ERR unknown command '" + cmd + "'"})
			continue
		}
		response := command.Handler(args)
		if command.IsWrite && response.Type != resp.RespError {
			slog.Info("Saving into AOF", "message", message)
			s.aof.Write(message)
		}
		writer.Write(response)
	}
	slog.Info("Disconnected", "remoteAddr", conn.RemoteAddr().String())
}
