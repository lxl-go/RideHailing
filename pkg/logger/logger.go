package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 日志配置（对标文档 5.5 节）
type Config struct {
	Level      string   // debug / info / warn / error
	Encoding   string   // json / console
	OutputPaths []string
	Service    string
	LogDir     string // 日志目录，空则 stdout
}

// New 创建 zap.Logger（对标文档日志打印统一规范）
func New(cfg Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.Set(cfg.Level); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.StacktraceKey = "stacktrace"

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var cores []zapcore.Core

	// 写入文件（按天切割）
	if cfg.LogDir != "" {
		fileWriter := getFileWriter(cfg.LogDir, cfg.Service)
		cores = append(cores, zapcore.NewCore(encoder, fileWriter, level))
	} else {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger = logger.With(zap.String("service", cfg.Service))

	return logger, nil
}

func getFileWriter(logDir, service string) zapcore.WriteSyncer {
	_ = os.MkdirAll(logDir+"/"+service, 0755)

	w := &rotateWriter{
		dir:       logDir + "/" + service,
		prefix:    service,
		maxSizeMB: 100,
		maxAgeDay: 30,
	}

	return zapcore.AddSync(w)
}

// rotateWriter 简易按天轮转
type rotateWriter struct {
	dir       string
	prefix    string
	maxSizeMB int
	maxAgeDay int
	current   *os.File
	bytes     int64
}

func (w *rotateWriter) Write(p []byte) (n int, err error) {
	if w.current == nil {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err = w.current.Write(p)
	w.bytes += int64(n)
	if w.bytes >= int64(w.maxSizeMB)*1024*1024 {
		w.current.Close()
		w.current = nil
		w.bytes = 0
	}
	return
}

func (w *rotateWriter) Sync() error {
	if w.current != nil {
		return w.current.Sync()
	}
	return nil
}

func (w *rotateWriter) rotate() error {
	now := time.Now().Format("2006-01-02")
	path := w.dir + "/" + w.prefix + "." + now + ".log"
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.current = f
	w.bytes = 0
	return nil
}
