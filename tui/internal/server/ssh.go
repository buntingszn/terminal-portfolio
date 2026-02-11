package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	"github.com/buntingszn/terminal-portfolio/tui/internal/analytics"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app/sections"
	"github.com/buntingszn/terminal-portfolio/tui/internal/config"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// SSHServer wraps a Wish SSH server that serves the Bubble Tea TUI.
type SSHServer struct {
	server      *ssh.Server
	logger      *slog.Logger
	content     *content.Content
	cfg         *config.Config
	analytics   *analytics.Logger
	maxSessions int64
	active      atomic.Int64
}

// New creates a new SSH server configured with Wish and Bubble Tea middleware.
func New(cfg *config.Config, c *content.Content) (*SSHServer, error) {
	al, err := analytics.NewLogger(cfg.AnalyticsFile)
	if err != nil {
		return nil, fmt.Errorf("create analytics logger: %w", err)
	}

	s := &SSHServer{
		logger:      slog.Default(),
		content:     c,
		cfg:         cfg,
		analytics:   al,
		maxSessions: int64(cfg.MaxSessions),
	}

	var srv *ssh.Server

	addr := fmt.Sprintf("%s:%d", cfg.SSHHost, cfg.SSHPort)

	if cfg.IdleTimeout > 0 {
		// Apply SSH-level idle timeout alongside the standard options.
		srv, err = wish.NewServer(
			wish.WithAddress(addr),
			wish.WithHostKeyPath(".ssh/terminal_portfolio_ed25519"),
			wish.WithIdleTimeout(cfg.IdleTimeout),
			wish.WithMiddleware(
				s.recoveryMiddleware(),
				s.sessionMiddleware(),
				bm.MiddlewareWithColorProfile(s.teaHandler, termenv.TrueColor),
			),
		)
	} else {
		// Idle timeout disabled (0); omit WithIdleTimeout entirely.
		srv, err = wish.NewServer(
			wish.WithAddress(addr),
			wish.WithHostKeyPath(".ssh/terminal_portfolio_ed25519"),
			wish.WithMiddleware(
				s.recoveryMiddleware(),
				s.sessionMiddleware(),
				bm.MiddlewareWithColorProfile(s.teaHandler, termenv.TrueColor),
			),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("create SSH server: %w", err)
	}

	s.server = srv
	return s, nil
}

// teaHandler returns a new Bubble Tea model for each SSH session.
func (s *SSHServer) teaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	theme := app.DarkTheme()
	m := app.New(s.content,
		sections.NewHomeSection(s.content, theme),
		sections.NewWorkSection(s.content, theme),
		sections.NewCVSection(s.content, theme),
		sections.NewLinksSection(s.content, theme),
	)
	// Wire idle timeout warning into the Bubbletea model so users
	// receive a 1-minute warning before the SSH idle disconnect.
	m = m.SetIdleTimeout(s.cfg.IdleTimeout)

	// Generate a short session ID and extract the visitor's IP for analytics.
	sid := strconv.FormatInt(time.Now().UnixMilli(), 36)
	remoteAddr := sess.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		ip = remoteAddr
	}

	s.analytics.Log(analytics.Event{
		Timestamp: time.Now(),
		SessionID: sid,
		Type:      analytics.EventSessionStart,
		IP:        ip,
	})
	m = m.SetAnalytics(s.analytics, sid, ip)

	opts := bm.MakeOptions(sess)
	opts = append(opts, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return m, opts
}

// recoveryMiddleware catches panics in SSH session handlers, logs them,
// and sends a user-friendly error message before closing the session.
func (s *SSHServer) recoveryMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("SSH session panic recovered",
						"panic", fmt.Sprintf("%v", r),
						"remote_addr", sess.RemoteAddr().String(),
					)
					_, _ = fmt.Fprintln(sess, "\r\nAn unexpected error occurred. Please reconnect.")
					_ = sess.Exit(1)
				}
			}()
			next(sess)
		}
	}
}

// sessionMiddleware returns Wish middleware that handles connection limits
// and session lifecycle logging.
func (s *SSHServer) sessionMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			remoteAddr := sess.RemoteAddr().String()
			ip, _, err := net.SplitHostPort(remoteAddr)
			if err != nil {
				ip = remoteAddr
			}

			logger := s.logger.With(
				"remote_addr", remoteAddr,
				"user", sess.User(),
				"ip", ip,
			)

			// Check global connection limit.
			current := s.active.Add(1)
			defer s.active.Add(-1)

			if current > s.maxSessions {
				logger.Warn("SSH connection rejected: at capacity",
					"active", current,
					"max", s.maxSessions,
				)
				_, _ = fmt.Fprintln(sess, "Server is at capacity. Please try again later.")
				_ = sess.Exit(1)
				return
			}

			logger.Info("SSH session started", "active_sessions", current)

			// Run the next handler (Bubble Tea).
			next(sess)

			logger.Info("SSH session ended")
		}
	}
}

// Start begins listening for SSH connections. This method blocks until
// the server is shut down or an error occurs.
func (s *SSHServer) Start() error {
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.server.Addr, err)
	}
	s.logger.Info("SSH server listening", "addr", ln.Addr().String())
	return s.server.Serve(ln)
}

// Shutdown gracefully shuts down the SSH server.
func (s *SSHServer) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	_ = s.analytics.Close()
	return err
}

// ActiveSessions returns the number of currently active sessions.
func (s *SSHServer) ActiveSessions() int64 {
	return s.active.Load()
}
