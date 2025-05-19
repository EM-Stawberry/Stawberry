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

	// Получаем имя файла и директорию для определения типа лога
	fileName := ""
	dirName := ""
	if entry.Caller.Defined {
		fileName = filepath.Base(entry.Caller.File)
		dirName = filepath.Base(filepath.Dir(entry.Caller.File))
	}
	
	// Определяем тип лога
	isSystemLog := strings.Contains(entry.Message, "Route:") || 
		strings.Contains(entry.Message, "Server started") || 
		strings.HasSuffix(fileName, "logger.go") ||
		dirName == "middleware" ||
		(entry.Caller.Defined && strings.Contains(entry.Caller.Function, "middleware"))
	
	// Если это лог от сервиса (не от Gin и не от middleware)
	if !isSystemLog && entry.Caller.Defined {
		// Получаем имя пакета из пути для сервисных логов
		filePath := entry.Caller.File
		fileName := filepath.Base(filePath)
		pkgPath := filepath.Dir(filePath)
		pkgName := filepath.Base(pkgPath)
		
		// Форматируем с выравниванием для сервисных логов
		fmt.Fprintf(buf, "%s%-25s%s ", colorDate, fmt.Sprintf("%s/%s:%d", pkgName, fileName, entry.Caller.Line), colorReset)
		fmt.Fprintf(buf, "%s%-30s%s", colorMsg, entry.Message, colorReset)
	} else {
		// Стандартное форматирование для Gin и middleware логов
		fmt.Fprintf(buf, "%s%s%s", colorMsg, entry.Message, colorReset)
		
		if entry.Caller.Defined && dirName != "middleware" && !strings.HasSuffix(fileName, "logger.go") {
			fmt.Fprintf(buf, " %s%s:%d%s", colorDate, fileName, entry.Caller.Line, colorReset)
		}
	}

	if len(fields) > 0 {
		tempEnc := zapcore.NewMapObjectEncoder()
		for _, field := range fields {
			field.AddTo(tempEnc)
		}

		if !isSystemLog && len(tempEnc.Fields) > 0 {
			// Красивый формат JSON для сервисных логов
			fmt.Fprintf(buf, " %s{", colorInfo)
			first := true
			for k, v := range tempEnc.Fields {
				if !first {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%q: %q", k, v)
				first = false
			}
			fmt.Fprintf(buf, "}%s", colorReset)
		} else {
			// Стандартный формат для логов Gin и middleware
			for k, v := range tempEnc.Fields {
				fmt.Fprintf(buf, " %s%s=%v%s", colorMsg, k, v, colorReset)
			}
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
