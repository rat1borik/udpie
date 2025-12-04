package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"udpie/internal/config"
	"udpie/internal/handler"
	"udpie/internal/service/signaller"
	"udpie/pkg/logutils"
)

const (
	shutdownTimeout = 30 * time.Second
)

func main() {
	// Load configuration
	cfg, err := config.LoadSignallerConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	err = initLogger(&cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Start server
	startServer(cfg)
}

func startServer(cfg *config.SignallerConfig) {
	producerService := signaller.NewProducerService()
	fileService := signaller.NewFileService()

	appRouter := handler.NewRouter(producerService, fileService)
	fastRouter := router.New()
	appRouter.SetupRoutes(fastRouter)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logutils.WithFields(logutils.Fields{
		"host": cfg.Server.Host,
		"port": cfg.Server.Port,
	}).Info("Signaller server starting")

	server := &fasthttp.Server{
		Handler:            fastRouter.Handler,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        30 * time.Second,
		CloseOnShutdown:    true,
		DisableKeepalive:   false,
		TCPKeepalive:       true,
		TCPKeepalivePeriod: 30 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logutils.WithError(err).Fatal("Failed to create listener")
	}

	serverErrChan := make(chan error, 1)
	go func() {
		if err := server.Serve(ln); err != nil {
			serverErrChan <- err
		}
	}()

	logutils.Info("Server started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		if err != nil {
			logutils.WithError(err).Error("Server error")
		}
	case sig := <-sigChan:
		gracefulShutdown(ln, serverErrChan, shutdownTimeout, sig)
	}
}

func gracefulShutdown(ln net.Listener, serverErrChan chan error, timeout time.Duration, sig os.Signal) {
	logutils.WithField("signal", sig.String()).Info("Received shutdown signal, initiating graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := ln.Close(); err != nil {
		logutils.WithError(err).Error("Error closing listener")
	} else {
		logutils.Info("Listener closed, no longer accepting new connections")
	}

	// Wait for active connections to finish or timeout
	done := make(chan struct{})
	go func() {
		// Wait for server to finish processing existing connections
		// The server will stop when all active connections are closed
		select {
		case err := <-serverErrChan:
			if err != nil {
				// Expected error when listener is closed
				logutils.Debug("Server stopped after listener close")
			}
		case <-shutdownCtx.Done():
			// Timeout reached
		}
		close(done)
	}()

	select {
	case <-done:
		logutils.Info("Server shutdown completed successfully")
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			logutils.Warn("Shutdown timeout exceeded, some connections may not have finished gracefully")
		}
	}
}

func initLogger(cfg *config.LogConfig) error {
	logConfig := logutils.LogConfig{
		Level:      cfg.Level,
		File:       cfg.File,
		Format:     cfg.Format,
		OutputFile: cfg.OutputFile,
		OutputStd:  cfg.OutputStd,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	err := logutils.SetupLogger(&logConfig)
	if err != nil {
		return err
	}

	logutils.Info("Logger initialized successfully")
	return nil
}
