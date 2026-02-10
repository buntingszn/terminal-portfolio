package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/buntingszn/terminal-portfolio/tui/internal/config"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
	"github.com/buntingszn/terminal-portfolio/tui/internal/server"
)

func main() {
	// Force true-color rendering on the global lipgloss default renderer.
	// This server process runs headless (no TTY), so termenv auto-detects
	// Ascii (no colors). All clients connect through ttyd/xterm.js or modern
	// terminals that support full 24-bit color.
	lipgloss.DefaultRenderer().SetColorProfile(termenv.TrueColor)
	lipgloss.DefaultRenderer().SetHasDarkBackground(true)

	// Load configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// Set up structured logging.
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// Log startup info.
	logger.Info("starting terminal-portfolio",
		"ssh_port", cfg.SSHPort,
		"data_dir", cfg.DataDir,
		"max_sessions", cfg.MaxSessions,
	)

	// Load content from JSON data files.
	c, err := content.LoadAll(cfg.DataDir)
	if err != nil {
		logger.Error("failed to load content", "err", err)
		os.Exit(1)
	}

	// Create SSH server.
	srv, err := server.New(cfg, c)
	if err != nil {
		logger.Error("failed to create SSH server", "err", err)
		os.Exit(1)
	}

	// Start server in a goroutine.
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("SSH server error", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGINT or SIGTERM for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("shutdown signal received", "signal", sig.String())

	// Graceful shutdown with 10-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}

	logger.Info("server stopped")
}
