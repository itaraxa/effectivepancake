package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Адаптер для zap
type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(levelString string) (logger *ZapLogger, err error) {
	var level zapcore.Level
	switch levelString {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level), // Уровень логирования
		Development: false,                       // Режим разработки (влияет на формат)
		Encoding:    "console",                   // Формат вывода "console" или "json"
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",                        // Ключ для вывода времени
			LevelKey:       "level",                       // Ключ для вывода уровня
			NameKey:        "logger",                      // Ключ для имени логгера
			CallerKey:      "caller",                      // Ключ для вывода информации о месте вызова
			MessageKey:     "msg",                         // Ключ для вывода сообщения
			StacktraceKey:  "stacktrace",                  // Ключ для стектрейса (выводится при ошибках)
			LineEnding:     zapcore.DefaultLineEnding,     // Конец строки (по умолчанию)
			EncodeLevel:    zapcore.CapitalLevelEncoder,   // Уровень логов (заглавные INFO, DEBUG и т.д.)
			EncodeTime:     zapcore.ISO8601TimeEncoder,    // Формат времени (ISO8601)
			EncodeDuration: zapcore.StringDurationEncoder, // Формат длительности
			EncodeCaller:   zapcore.ShortCallerEncoder,    // Формат вызова (короткий путь к файлу)
		},
		OutputPaths:      []string{"stdout"}, // Куда выводить логи (в данном случае в консоль)
		ErrorOutputPaths: []string{"stderr"}, // Куда выводить ошибки
	}

	l, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return
	}

	return &ZapLogger{logger: l}, nil
}

// Преобразуем интерфейсные поля в zap.Field и вызываем соответствующий метод ZapLogger.logger
func (zl *ZapLogger) Info(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	zl.logger.Info(msg, zapFields...)
}

func (zl *ZapLogger) Debug(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	zl.logger.Debug(msg, zapFields...)
}

func (zl *ZapLogger) Error(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	zl.logger.Error(msg, zapFields...)
}

func (zl *ZapLogger) Fatal(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	zl.logger.Fatal(msg, zapFields...)
}

func (zl *ZapLogger) Sync() {
	_ = zl.logger.Sync()

}

// Вспомогательная функция для преобразования полей в zap.Field
// для поля типа time.duration отдельная проверка
func convertToZapFields(fields ...interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				if key != `duration` {
					zapFields = append(zapFields, zap.Any(key, fields[i+1]))
				} else {
					if duration, ok2 := fields[i+1].(time.Duration); ok2 {
						zapFields = append(zapFields, zap.Duration(key, duration))
					}
				}
			}
		}
	}
	return zapFields
}
