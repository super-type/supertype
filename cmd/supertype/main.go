package main

import (
	"net/http"

	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/http/rest"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage/dynamo"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Set up logging using Uber's Zap logger
func initLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}

func main() {
	// Initialize storage
	persistentStorage := new(dynamo.Storage)

	// Initialize logger
	loggerManager := initLogger()
	zap.ReplaceGlobals(loggerManager)
	defer loggerManager.Sync() // flushes buffer, if any
	logger := loggerManager.Sugar()

	// Initialize services
	authenticator := authenticating.NewService(persistentStorage)
	dashboard := dashboard.NewService(persistentStorage)
	producing := producing.NewService(persistentStorage)
	consuming := consuming.NewService(persistentStorage)

	// Initialize routers and startup server
	httpRouter := rest.Router(authenticator, producing, consuming, dashboard)
	logger.Info("Starting HTTP server on port 5000...")
	logger.Fatal(http.ListenAndServe(":5000", httpRouter))
}
