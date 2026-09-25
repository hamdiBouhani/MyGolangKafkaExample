package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/codec"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/config"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/producer"
)

func main() {
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		// Can't use logger yet; write to stderr.
		_, _ = os.Stderr.WriteString("config error: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 2. Logger
	log, err := buildLogger(cfg.LogLevel)
	if err != nil {
		_, _ = os.Stderr.WriteString("logger error: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting kafka producer",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
		zap.String("schema_registry", cfg.SchemaRegistryURL),
	)

	// 3. Producer service
	svc, err := producer.New(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize producer", zap.Error(err))
	}

	// 4. Metrics server (separate goroutine, its own lifecycle)
	metricsSrv := startMetricsServer(cfg.MetricsAddr, log)

	// 5. Root context cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 6. Produce loop
	if err := run(ctx, cfg, svc, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("produce loop ended with error", zap.Error(err))
	}

	// 7. Graceful shutdown
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := svc.Close(); err != nil {
		log.Error("producer close failed", zap.Error(err))
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("metrics server shutdown failed", zap.Error(err))
	}
	log.Info("shutdown complete")
}

func run(ctx context.Context, cfg *config.Config, svc *producer.Service, log *zap.Logger) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sent := 0
	for {
		if cfg.NumMessages > 0 && sent >= cfg.NumMessages {
			log.Info("reached NUM_MESSAGES, exiting", zap.Int("sent", sent))
			return nil
		}

		select {
		case <-ctx.Done():
			log.Info("context cancelled, stopping produce loop", zap.Int("sent", sent))
			return ctx.Err()
		case <-ticker.C:
			evt := codec.UserEvent{
				UserID:    "user-123",
				EventType: "LOGIN",
				Timestamp: time.Now().Unix(),
			}
			if err := svc.ProduceUserEvent(ctx, evt); err != nil {
				// Log and continue; a single failure shouldn't kill the service.
				log.Error("produce failed", zap.Error(err))
				continue
			}
			sent++
		}
	}
}

func buildLogger(level string) (*zap.Logger, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	zcfg := zap.NewProductionConfig()
	zcfg.Level = zap.NewAtomicLevelAt(lvl)
	zcfg.EncoderConfig.TimeKey = "ts"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zcfg.Build()
}

func startMetricsServer(addr string, log *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		log.Info("metrics server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server error", zap.Error(err))
		}
	}()
	return srv
}
