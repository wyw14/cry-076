package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/config"
	"github.com/wyw14/cry-076/internal/platform"
	"github.com/wyw14/cry-076/internal/repository/postgres"
	transport "github.com/wyw14/cry-076/internal/transport/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := buildLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	dbCtx, cancelDB := context.WithTimeout(rootCtx, cfg.Database.Timeout)
	db, err := postgres.Open(dbCtx, cfg.Database.URL)
	cancelDB()
	if err != nil {
		logger.Fatal("database unavailable", zap.Error(err))
	}
	defer db.Close()

	clock := platform.UTCClock{}
	ids := platform.RandomIDs{}
	notifier := &platform.ResumeNotifier{}
	files := platform.LocalFileStore{Root: cfg.Storage.Root}
	renderer := platform.OfflineRenderer{}

	templates := postgres.TemplateRepository{DB: db}
	drafts := postgres.DraftRepository{DB: db}
	snapshots := postgres.SnapshotRepository{DB: db}
	mappings := postgres.MappingRepository{DB: db}
	profiles := postgres.ProfileRepository{DB: db}
	privacy := postgres.PrivacyRepository{DB: db}
	exports := postgres.ExportRepository{DB: db}
	feedback := postgres.FeedbackRepository{DB: db}
	attachments := postgres.AttachmentRepository{DB: db}
	audits := postgres.AuditRepository{DB: db}

	handlers := transport.Handlers{
		Templates:   application.NewTemplateService(templates, feedback, audits, notifier, clock, ids),
		Drafts:      application.NewDraftService(drafts, snapshots, templates, audits, db, clock, ids),
		Switches:    application.NewSwitchService(drafts, templates, mappings, snapshots, audits, db, clock, ids),
		Consistency: application.NewConsistencyService(drafts, templates, privacy),
		Profiles:    application.NewProfileService(profiles, audits, clock, ids),
		Privacy:     application.NewPrivacyService(privacy, audits, clock, ids),
		Exports:     application.NewExportService(drafts, snapshots, templates, privacy, exports, files, renderer, audits, clock, ids),
		Attachments: application.NewAttachmentService(attachments, privacy, files, audits, clock, ids, cfg.Storage.MaxUploadBytes, cfg.Storage.AllowedTypes),
		Feedback:    application.NewFeedbackService(feedback, templates, audits, notifier, clock, ids),
		Audits:      application.NewAuditService(audits),
	}

	ready := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return db.Ready(ctx) == nil
	}
	router := transport.NewRouter(logger, cfg.HTTP.AllowedOrigins, ready, handlers)
	server := &http.Server{Addr: cfg.HTTP.Address, Handler: router, ReadHeaderTimeout: cfg.HTTP.ReadTimeout, ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout, IdleTimeout: 60 * time.Second}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server started", zap.String("address", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		logger.Error("server failed", zap.Error(err))
		stop()
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
		_ = server.Close()
	}
}

func buildLogger(cfg config.Config) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Logging.Level)); err != nil {
		return nil, err
	}
	loggerCfg := zap.NewProductionConfig()
	if cfg.Logging.Development {
		loggerCfg = zap.NewDevelopmentConfig()
	}
	loggerCfg.Level = zap.NewAtomicLevelAt(level)
	loggerCfg.OutputPaths = []string{"stdout"}
	loggerCfg.ErrorOutputPaths = []string{"stderr"}
	loggerCfg.EncoderConfig.TimeKey = "timestamp"
	loggerCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return loggerCfg.Build()
}
