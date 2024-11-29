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
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
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
		fx.Provide(NewLoggerProvider),
		fx.Provide(NewLogger),
		fx.Invoke(SetupOtelSdk),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			fxLogger := &fxevent.ZapLogger{Logger: logger}
			fxLogger.UseLogLevel(zapcore.DebugLevel)
			return fxLogger
		}),
		fx.Provide(
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
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func NewLoggerProvider() (provider *log.LoggerProvider, err error) {
	return otel.NewLoggerProvider(context.Background())
}

func NewLogger(lgp *log.LoggerProvider) (*zap.Logger, error) {
	var cores []zapcore.Core

	if shouldLogLoki := common.DefaultGetEnvBool("LOG_LOKI", false); shouldLogLoki {
		lokiCore := otelzap.NewCore("go-seed", otelzap.WithLoggerProvider(lgp))
		cores = append(cores, lokiCore)
	}

	if shouldLogStdout := common.DefaultGetEnvBool("LOG_STDOUT", true); len(cores) == 0 || shouldLogStdout {
		stdoutCore := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
		cores = append(cores, stdoutCore)
	}

	return zap.New(zapcore.NewTee(cores...)), nil
}

func SetupOtelSdk(lc fx.Lifecycle, logger *zap.Logger, lgp *log.LoggerProvider) error {
	shutdownFunc, err := otel.SetupOTelSdk(lgp)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down OpenTelemetry")
			return shutdownFunc(ctx)
		},
	})

	return err
}

func NewHttpServer(lc fx.Lifecycle, handler http.Handler, logger *zap.Logger) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", (common.MustGetEnv("PORT"))),
		Handler: handler,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
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

	return mux
}

func asRoute(handlerFactory any) any {
	return fx.Annotate(
		handlerFactory,
		fx.ResultTags(`group:"routes"`),
	)
}
