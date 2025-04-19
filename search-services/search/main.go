package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/adapters/db"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/adapters/timer"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to the configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debag message are enable")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		return
	}

	words, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("failed create Words client", "error", err)
		return
	}

	index := core.NewIndex(log, storage)

	timer := timer.New(cfg.TTL, index)
	go func() {
		timer.Start(ctx)
	}()

	searcher := core.NewService(log, storage, words, index)

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Error("failed to listen", "error", err)
		return
	}

	s := grpc.NewServer()
	searchpb.RegisterSearchServer(s, searchgrpc.NewServer(searcher))
	reflection.Register(s)

	go func() {
		<-ctx.Done()
		log.Info("server stopped")
		s.GracefulStop()
	}()

	if err := s.Serve(listener); err != nil {
		log.Error("failed to serve", "error", err)
		return
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
