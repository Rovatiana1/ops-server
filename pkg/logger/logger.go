package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

// Init initialises the global logger.
// level: "debug" | "info" | "warn" | "error"
func Init(level string, isDev bool) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	var encoder zapcore.Encoder
	if isDev {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	global = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// L returns the global logger. Panics if Init was not called.
func L() *zap.Logger {
	if global == nil {
		panic("logger not initialised: call logger.Init() first")
	}
	return global
}

// S returns the sugared global logger.
func S() *zap.SugaredLogger {
	return L().Sugar()
}

// Sync flushes any buffered log entries.
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}
