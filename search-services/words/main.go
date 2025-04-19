package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	wordspb "yadro.com/course/proto/words"
	wordsgrpc "yadro.com/course/words/adapters/grpc"
	"yadro.com/course/words/adapters/stemming"
	"yadro.com/course/words/core"
)

const (
	maxShutdownTime = 5 * time.Second
)

type Config struct {
	Port string `yaml:"port" env:"WORDS_GRPC_PORT" env-default:"11111"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	log := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	))

	if err := run(cfg, log); err != nil {
		log.Error("failed to run", "error", err)
		os.Exit(1)
	}
}

func run(cfg Config, log *slog.Logger) error {

	// adapter for stemmer
	stemmer := stemming.Snowball{}

	// core service
	words := core.NewWords(stemmer)

	// grpc adapter
	server := wordsgrpc.New(words)

	// server and shutdown logic
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen port %s: %v", cfg.Port, err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, server)
	reflection.Register(s)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("starting server", "port", cfg.Port)
		if err = s.Serve(listener); err != nil {
			log.Error("faied to serve", "error", err)
			// notify shutdown code if needed
			cancel()
		}
	}()

	// got failed serve or terminate signal
	<-ctx.Done()

	// force after some time
	timer := time.AfterFunc(maxShutdownTime, func() {
		log.Info("forcing server stop")
		s.Stop()
	})
	defer timer.Stop()

	log.Info("starting graceful stop")
	s.GracefulStop()
	log.Info("server stopped")

	return err
}
