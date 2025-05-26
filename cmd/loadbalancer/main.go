package main

import (
	"cloud-test/internal"
	"cloud-test/internal/configuration"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	path, err := getPathToConfigurationFileFromFlag()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	cfg, err := configuration.Load(path)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	if cfg.LoggerPath != "" {
		logFile, err := os.OpenFile(cfg.LoggerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			return
		}
		defer logFile.Close()

		fileHandler := slog.NewTextHandler(logFile, nil)

		logger := slog.New(fileHandler)
		slog.SetDefault(logger)
	}

	quitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := internal.Run(cfg, quitCtx); err != nil {
		slog.Error("failed to run server", "error", err)
	}
}

func getPathToConfigurationFileFromFlag() (string, error) {
	fmt.Println(os.Args)
	configurationPath := flag.String("configuration", "", "path to configuration file")
	flag.StringVar(configurationPath, "c", "", "path to configuration file")
	flag.Parse()

	if *configurationPath == "" {
		return "", fmt.Errorf("configuration file path is not provided")
	}

	_, err := os.Stat(*configurationPath)
	if err != nil {
		return "", fmt.Errorf("configuration file does not exist")
	}

	return *configurationPath, nil
}
