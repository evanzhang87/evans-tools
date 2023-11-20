package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	bootstrapLogPath = "bootstrap.log"
	runtimeLogPath   = "runtime.log"
)

type LogConfig struct {
	Filename   string `json:"filename" yaml:"filename"`
	Level      string `json:"level" yaml:"level"`
	MaxSize    int    `json:"maxsize" yaml:"maxsize"`
	MaxAge     int    `json:"maxage" yaml:"maxage"`
	MaxBackups int    `json:"maxbackups" yaml:"maxbackups"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

var (
	Logger      *zap.SugaredLogger
	writeSyncer zapcore.WriteSyncer
	encoder     zapcore.Encoder
	core        zapcore.Core
	logger      *zap.Logger
)

func init() {
	writeSyncer = getLogWriter(bootstrapLogPath)
	encoder = getEncoder()
	core = zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	logger = zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

func InitLogger(config *LogConfig) {
	_ = Logger.Sync()
	lumberJackLogger := &lumberjack.Logger{
		Filename:   runtimeLogPath,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   false,
	}
	if config.Filename != "" {
		lumberJackLogger.Filename = config.Filename
	}
	if config.MaxSize > 0 {
		lumberJackLogger.MaxSize = config.MaxSize
	}
	if config.MaxBackups > 0 {
		lumberJackLogger.MaxBackups = config.MaxBackups
	}
	if config.MaxAge > 0 {
		lumberJackLogger.MaxAge = config.MaxAge
	}
	if config.Compress {
		lumberJackLogger.Compress = config.Compress
	}

	writeSyncer = zapcore.AddSync(lumberJackLogger)
	logLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		Logger.Warnf("parse log level err: %v", err)
		logLevel = zapcore.InfoLevel
	}
	core = zapcore.NewCore(encoder, writeSyncer, logLevel)
	logger = zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(path string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func GetLogger() *zap.SugaredLogger {
	return Logger
}
