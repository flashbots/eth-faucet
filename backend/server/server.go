package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flashbots/eth-faucet/config"
	"github.com/flashbots/eth-faucet/httplogger"
	"github.com/flashbots/eth-faucet/logutils"
	"github.com/flashbots/eth-faucet/ratelimiter"
	"github.com/flashbots/eth-faucet/txbuilder"
	"go.uber.org/zap"
)

var (
	ErrRatelimiterFailedToInitialise        = errors.New("failed to initialise rate-limiter")
	ErrTransactionBuilderFailedToInitialise = errors.New("failed to initialise transactions builder")
)

type Server struct {
	cfg         *config.Config
	log         *zap.Logger
	ratelimiter *ratelimiter.RateLimiter
	txbuilder   *txbuilder.TxBuilder
}

func New(cfg *config.Config) (*Server, error) {
	ratelimiter, err := ratelimiter.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRatelimiterFailedToInitialise, err)
	}

	txbuilder, err := txbuilder.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTransactionBuilderFailedToInitialise, err)
	}

	return &Server{
		cfg:         cfg,
		log:         zap.L(),
		ratelimiter: ratelimiter,
		txbuilder:   txbuilder,
	}, nil
}

func (s *Server) Run() error {
	l := s.log
	ctx := logutils.ContextWithLogger(context.Background(), l)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fund", s.handleFund)
	mux.HandleFunc("/api/info", s.handleInfo)
	handler := httplogger.Middleware(l, mux)

	srv := &http.Server{
		Addr:              s.cfg.Server.ListenAddress,
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		terminator := make(chan os.Signal, 1)
		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)
		stop := <-terminator

		l.Info("Stop signal received; shutting down...", zap.String("signal", stop.String()))

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			l.Error("HTTP server shutdown failed",
				zap.Error(err),
			)
		}
	}()

	l.Info("Starting up faucet server...",
		zap.String("server_listen_address", s.cfg.Server.ListenAddress),
	)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("Faucet server failed", zap.Error(err))
	}
	l.Info("Server is down")

	return nil
}
