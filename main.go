package main

import (
	"context"
	"gokit-seed/internal/common"
	"gokit-seed/internal/test"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Route interface {
	http.Handler
	Pattern() string
}

func main() {
	if os.Getenv("GO_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			logger, _ := zap.NewProduction()
			logger.Error("failed to load dotenv", zap.Error(err))
			panic(err)
		}
	}

	fx.New(
		fx.Provide(
			NewLogger,
			NewHttpServer,
			fx.Annotate(
				NewMuxServer,
				fx.ParamTags(`group:"routes"`),
			),
			// Add more services here
      test.NewTestService,
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

func NewHttpServer(lc fx.Lifecycle, mux *http.ServeMux, logger *zap.Logger) *http.Server {
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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

func NewMuxServer(routes []Route, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range routes {
		pattern := route.Pattern()
		prefix := pattern[:len(pattern)-1]
		handler := common.LoggingHandler(logger, route)
		mux.Handle(pattern, http.StripPrefix(prefix, handler))
	}

	return mux
}

func asRoute(handlerFactory any) any {
	return fx.Annotate(
		handlerFactory,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}