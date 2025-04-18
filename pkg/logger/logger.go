package logger

import "go.uber.org/zap"

var (
	envLocal = "local"
	envProd  = "prod"
)

// SetupLogger настраивает логгер в зависимости от окружения (`local` или `prod`).
// Использует `zap.NewDevelopment()` для разработки с уровнем `DebugLevel` и читаемым выводом,
// и `zap.NewProduction()` для продакшн-среды с уровнем `InfoLevel` и сжатым JSON выводом.
func SetupLogger(env string) *zap.Logger {
	var log *zap.Logger
	var err error

	switch env {
	case envLocal:
		log, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	case envProd:
		log, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	}

	return log
}
