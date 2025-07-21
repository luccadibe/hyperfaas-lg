package main

import (
	"flag"
	"lg/internal"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/goforj/godump"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	config := flag.String("config", "workload_config.yaml", "config file")
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()

	logLevelInt := getLogLevel(*logLevel)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(logLevelInt),
	}))

	controller := internal.NewController(
		logger,
		internal.WithConfigFile(*config),
	)
	godump.Dump(controller.Config.Workload)
	controller.Run()
}

func getLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
