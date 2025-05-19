package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zuzaaa-dev/stawberry/config"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	colorDate  = "\033[38;5;246m"
	colorMsg   = "\033[38;5;15m"
	colorDebug = "\033[36m"
	colorInfo  = "\033[32m"
	colorWarn  = "\033[33m"
	colorError = "\033[31m"
	colorReset = "\033[0m"
)

type SimpleEncoder struct {
	zapcore.Encoder
	pool buffer.Pool
}

func (e *SimpleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := e.pool.Get()

	timeStr := entry.Time.Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(buf, "%s%s%s ", colorDate, timeStr, colorReset)

	levelColor := getLevelColor(entry.Level)
	fmt.Fprintf(buf, "%s%-5s%s ", levelColor, strings.ToUpper(entry.Level.String()), colorReset)
	fmt.Fprintf(buf, "%s%s%s", colorMsg, entry.Message, colorReset)

	if entry.Caller.Defined {
		fileName := filepath.Base(entry.Caller.File)
		fmt.Fprintf(buf, " %s%s:%d%s", colorDate, fileName, entry.Caller.Line, colorReset)
	}

	if len(fields) > 0 {
		tempEnc := zapcore.NewMapObjectEncoder()
		for _, field := range fields {
			field.AddTo(tempEnc)
		}

		for k, v := range tempEnc.Fields {
			fmt.Fprintf(buf, " %s%s=%v%s", colorMsg, k, v, colorReset)
		}
	}

	buf.AppendByte('\n')
	return buf, nil
}

func getLevelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return colorDebug
	case zapcore.InfoLevel:
		return colorInfo
	case zapcore.WarnLevel:
		return colorWarn
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return colorError
	default:
		return colorMsg
	}
}

func getEncoder(isJSON bool) zapcore.Encoder {
	if isJSON {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return &SimpleEncoder{
		Encoder: zapcore.NewConsoleEncoder(encoderConfig),
		pool:    buffer.NewPool(),
	}
}

func SetupLogger(env string) *zap.Logger {
	var level zapcore.Level
	var isJSON bool

	switch env {
	case config.EnvDev, config.EnvTest:
		level = zap.DebugLevel
		isJSON = false
	case config.EnvProd:
		level = zap.InfoLevel
		isJSON = true
	default:
		level = zap.InfoLevel
		isJSON = false
	}

	core := zapcore.NewCore(
		getEncoder(isJSON),
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller())
}
