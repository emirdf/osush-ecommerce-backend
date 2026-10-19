package main

import (
	"os"
	"osush/internal/app"
	"osush/internal/log"
	zapLogger "osush/internal/log/zap"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	logger, cleanup, err := zapLogger.NewLogger(env)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer cleanup()

	logger.Info("starting application",
		log.String("env", env),
		log.String("version", "1.0"),
	)

	if err := app.Run(logger); err != nil {
		logger.Fatal("application failed to start",
			log.Error(err),
		)
	}
}
