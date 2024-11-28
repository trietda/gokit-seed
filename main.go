package main

import (
	"context"
	"fmt"
	"gokit-seed/internal/common"
	"gokit-seed/internal/otel"
	"gokit-seed/internal/test"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if os.Getenv("GO_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			logger, _ := zap.NewProduction()
			logger.Error("failed to load dotenv", zap.Error(err))
			panic(err)
		}
	}

	var (
		TEST_URL = common.GetEnv("TEST_URL")
	)

	fx.New(
		fx.Provide(
			NewLogger,
			NewHttpServer,
			fx.Annotate(
				NewMuxServer,
				fx.ParamTags(`group:"routes"`),
			),
			// Add more services here
			func() test.TestService {
				testService := test.NewTestService()
				testService = test.MakeProxyTestService(TEST_URL)(testService)
				return testService
			},

			// Add more routes here
			asRoute(test.MakeHandler),
		),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			fxLogger := &fxevent.ZapLogger{Logger: logger}
			fxLogger.UseLogLevel(zapcore.DebugLevel)
			return fxLogger
		}),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func NewLogger() (*zap.Logger, error) {
	switch os.Getenv("GO_ENV") {
	case "production":
		return zap.NewProduction()
	default:
		return zap.NewDevelopment()
	}
}

func NewHttpServer(lc fx.Lifecycle, handler http.Handler, logger *zap.Logger) *http.Server {
	var shutdownOtel func(context.Context) error
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", (common.MustGetEnv("PORT"))),
		Handler: handler,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			shutdownOtel, _ = otel.SetupOTelSDK(ctx)
			ln, err := net.Listen("tcp", server.Addr)

			if err != nil {
				return err
			}

			logger.Info(
				"server is listening",
				zap.String("addr", server.Addr),
			)
			go server.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("otel is shutting down")
			if err := shutdownOtel(ctx); err != nil {
				return err
			}

			logger.Info("server is shutting down")
			return server.Shutdown(ctx)
		},
	})

	return server
}

func NewMuxServer(routes []*common.RouteGroup, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	for _, route := range routes {
		path := route.Path + "/"
		handler := common.LoggingHandler(logger, route)
		mux.Handle(path, handler)
	}

	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

func asRoute(handlerFactory any) any {
	return fx.Annotate(
		handlerFactory,
		fx.ResultTags(`group:"routes"`),
	)
}
