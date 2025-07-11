package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// Log seviyesi
	logLevel := zapcore.InfoLevel

	// Özel encoder ayarları
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Konsola log yazacak sync
	consoleSync := zapcore.AddSync(os.Stdout)

	// Log dosyasına yazmak için dosya oluştur
	logFile, err := os.OpenFile("./pkg/log/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic("Log dosyası açılamadı: " + err.Error())
	}
	fileSync := zapcore.AddSync(logFile)

	// Her iki hedefe de yazacak olan Core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleSync, logLevel),
		zapcore.NewCore(encoder, fileSync, logLevel),
	)

	// Logger oluştur
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Global logger olarak ayarla
	zap.ReplaceGlobals(logger)
}
