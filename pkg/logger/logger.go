package logger

import (
	"github.com/zuzaaa-dev/stawberry/config"
	"go.uber.org/zap"
)

// SetupLogger настраивает логгер в зависимости от окружения (`dev`, `test`, `prod`).
// Использует `zap.NewDevelopment()` для разработки с уровнем `DebugLevel` и читаемым выводом,
// и `zap.NewProduction()` для продакшн-среды с уровнем `InfoLevel` и сжатым JSON выводом.
func SetupLogger(env string) *zap.Logger {
	var log *zap.Logger
	var err error

	switch env {
	case config.EnvDev:
		log, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	case config.EnvTest:
		log, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	case config.EnvProd:
		log, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	}

	return log
}
