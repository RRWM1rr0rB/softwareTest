package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	gRPCOrderService "github.com/WM1rr0rB8/contractsTest/gen/go/order_service/v1"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/apperror"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/core/closer"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/core/healthcheck"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/core/safe"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/metrics"
	psql "github.com/WM1rr0rB8/librariesTest/backend/golang/postgresql"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/tracing"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/utils/clock"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/utils/ident"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"software_test/internal/config"
	gRPCOrder "software_test/internal/controller/grpc/v1/order"
	orderHTTP "software_test/internal/controller/http/v1/order"
	"software_test/internal/dal/postgres"
	"software_test/internal/domain"
	domainOrderService "software_test/internal/domain/order/service"
	domainOrderStorage "software_test/internal/domain/order/storage"
	"software_test/internal/policy"
	policyOrder "software_test/internal/policy/order"
)

type Runner interface {
	Run(context.Context) error
}

type App struct {
	cfg        *config.Config
	gRPCServer *grpc.Server
	httpRouter *chi.Mux
	httpServer *http.Server

	metricsHTTTPServer *metrics.Server
	healthServer       *healthcheck.GRPCHealthServer

	policyOrder *policyOrder.Policy

	runners []Runner
}

func (a *App) AddRunner(runner Runner) {
	a.runners = append(a.runners, runner)
}

//nolint:funlen
func NewApp(ctx context.Context) (*App, error) {
	app := App{}

	cfg := config.GetConfig()
	app.cfg = cfg

	logger := logging.NewLogger(
		logging.WithLevel(cfg.App.LogLevel),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.L(ctx).Info("config loaded", "config", cfg)

	app.healthServer = healthcheck.NewGRPCHealthServer()

	// Init Trace Server.
	err := initTraceServer(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "initTraceServer")
	}

	if cfg.Metrics.Enabled {
		var metricsErr error

		app.metricsHTTTPServer, metricsErr = metrics.NewServer(metrics.NewConfig(
			metrics.WithHost(cfg.Metrics.Host),
			metrics.WithPort(cfg.Metrics.Port),
			metrics.WithReadTimeout(cfg.Metrics.ReadTimeout),
			metrics.WithWriteTimeout(cfg.Metrics.WriteTimeout),
			metrics.WithReadHeaderTimeout(cfg.Metrics.ReadHeaderTimeout),
		))
		if metricsErr != nil {
			return nil, errors.Wrap(metricsErr, "can't create metrics server")
		}

		closer.Add(app.metricsHTTTPServer)
	}

	// Init Postgres Client.
	postgresClient, err := app.initPostgresClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't create postgres Client")
	}

	uuidGenerator := ident.NewUUIDGenerator()
	defClock := clock.NewDefault()

	// Init storage and service.
	orderStorage := domainOrderStorage.NewStorage(postgresClient)
	orderService := domainOrderService.NewService(orderStorage)

	// Init policy.
	basePolicy := policy.NewBasePolicy(
		uuidGenerator,
		defClock,
	)

	app.policyOrder = policyOrder.NewPolicy(
		basePolicy,
		orderService,
		cfg,
	)
	// init gRPC controllers
	app.gRPCServer = app.initGRPCServer(ctx)

	// init HTTP router
	app.httpRouter = app.initHTTPRouter(ctx)

	return &app, nil
}

func (a *App) Run(ctx context.Context) error {
	// Run migrations.
	err := postgres.RunMigrations(&a.cfg.Postgres)
	if err != nil {
		return errors.Wrap(err, "migrations failed")
	}

	errGroup, _ := safe.WithContext(ctx)

	errGroup.Run(closer.CloseOnSignalContext(os.Kill, os.Interrupt))
	errGroup.Run(a.setupGRPCServer)
	errGroup.Run(a.setupHTTPServer)

	if a.cfg.Metrics.Enabled {
		errGroup.Run(a.metricsHTTTPServer.Run)
	}

	for _, r := range a.runners {
		errGroup.Run(r.Run)
	}

	logging.L(ctx).Info("application started")

	return errGroup.Wait()
}

func (a *App) setupGRPCServer(ctx context.Context) error {
	logging.L(ctx).Info(
		"gRPC server initializing",
		logging.StringAttr("host", a.cfg.GRPC.Host),
		logging.IntAttr("port", a.cfg.GRPC.Port),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.GRPC.Host, a.cfg.GRPC.Port))
	if err != nil {
		return errors.Wrap(err, "gRPC server listen error")
	}

	closer.Add(lis)

	if err = a.gRPCServer.Serve(lis); err != nil {
		return errors.Wrap(err, "gRPC server serve error")
	}

	return nil
}

func (a *App) setupHTTPServer(ctx context.Context) error {
	logging.L(ctx).Info(
		"HTTP server initializing",
		logging.StringAttr("host", a.cfg.HTTP.Host),
		logging.IntAttr("port", a.cfg.HTTP.Port),
		logging.DurationAttr("read_timeout", a.cfg.HTTP.ReadHeaderTimeout),
	)

	a.httpServer = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port),
		Handler:     a.httpRouter,
		ReadTimeout: a.cfg.HTTP.ReadHeaderTimeout,
	}

	closer.Add(a.httpServer)

	if err := a.httpServer.ListenAndServe(); err != nil {
		logging.L(ctx).With(logging.ErrAttr(err)).Error("HTTP server listen and serve error")
		return err
	}

	return nil
}

func (a *App) initGRPCServer(ctx context.Context) *grpc.Server {
	logging.L(ctx).Info(
		"gRPC server initializing",
		logging.StringAttr("host", a.cfg.GRPC.Host),
		logging.IntAttr("port", a.cfg.GRPC.Port),
	)

	recoveryHandler := grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "internal system error")
	})

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			apperror.GRPCUnaryInterceptor(domain.SystemCode),
			logging.WithTraceIDInLogger(),
			metrics.RequestDurationMetricUnaryServerInterceptor(fmt.Sprintf(
				"%s-%s-%s",
				a.cfg.App.Name,
				a.cfg.App.ID,
				a.cfg.App.Version,
			)),
			grpc_recovery.UnaryServerInterceptor(recoveryHandler),
		),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(recoveryHandler),
		),
	}

	serverOptions = append(serverOptions, tracing.WithAllTracing()...)

	gRPCServer := grpc.NewServer(
		serverOptions...,
	)
	reflection.Register(gRPCServer)

	grpc_health_v1.RegisterHealthServer(gRPCServer, a.healthServer.Server)

	a.healthServer.HealthCheck(ctx, a.cfg.GRPC.HealthCheckInterval)

	// Init controllers.
	gRPCOrderService.RegisterOrderServiceServer(gRPCServer,
		gRPCOrder.NewController(
			a.policyOrder,
		),
	)

	return gRPCServer
}

func initTraceServer(ctx context.Context, cfg *config.Config) error {
	if !cfg.Tracing.Enabled {
		logging.L(ctx).Info("tracing is disabled")

		return nil
	}

	_, err := tracing.New(
		tracing.WithHost(cfg.Tracing.Host),
		tracing.WithPort(strconv.Itoa(cfg.Tracing.Port)),
		tracing.WithServiceID(cfg.App.ID),
		tracing.WithServiceName(cfg.App.Name),
		tracing.WithServiceVersion(cfg.App.Version),
	)
	if err != nil {
		return errors.Wrap(err, "can't initialize tracing")
	}

	return nil
}

func (a *App) initHTTPRouter(_ context.Context) *chi.Mux {
	router := chi.NewRouter()

	router.Use(logging.Middleware)
	router.Use(tracing.Middleware)

	router.Use(metrics.RequestDurationMetricHTTPMiddleware)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	ordersHTTP := orderHTTP.NewController(
		a.policyOrder,
	)

	router.Post("/create_order", ordersHTTP.CreateOrder)

	return router
}

func (a *App) initPostgresClient(ctx context.Context) (*psql.Client, error) {
	logging.WithAttrs(
		ctx,
		logging.StringAttr("host", a.cfg.Postgres.Host),
		logging.IntAttr("port", a.cfg.Postgres.Port),
		logging.StringAttr("user", a.cfg.Postgres.User),
		logging.StringAttr("db", a.cfg.Postgres.Database),
		logging.StringAttr("password", "<REMOVED>"),
		logging.IntAttr("max-attempts", a.cfg.Postgres.MaxAttempt),
		logging.DurationAttr("max_delay", a.cfg.Postgres.MaxDelay),
	).Info("PostgreSQL initializing")

	pgDsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		a.cfg.Postgres.User,
		a.cfg.Postgres.Password,
		a.cfg.Postgres.Host,
		a.cfg.Postgres.Port,
		a.cfg.Postgres.Database,
	)

	if a.cfg.Postgres.Binary {
		pgDsn += "?sslmode=require"
	}

	postgresConfig, err := psql.NewConfig(
		pgDsn,
		a.cfg.Postgres.MaxAttempt,
		a.cfg.Postgres.MaxDelay,
		psql.WithBinaryExecMode(a.cfg.Postgres.Binary),
	)
	if err != nil {
		return nil, errors.Wrap(err, "psql.NewConfig")
	}

	pgClient, err := psql.NewClient(ctx, postgresConfig)
	if err != nil {
		return nil, errors.Wrap(err, "psql.NewClient")
	}

	closer.AddN(pgClient)

	return pgClient, nil
}
