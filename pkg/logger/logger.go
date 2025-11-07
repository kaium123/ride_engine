package logger

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc/metadata"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const TraceId = "trace_id"
const UserId = "user_id"
const UserName = "user_name"

var logger = logrus.New()

// DefaultLogger return configured default logger
func DefaultLogger() *logrus.Logger {
	return logger
}

// Fields wraps logrus.Fields, which is a map[string]interface{}
type Fields logrus.Fields

// SetLogLevel ...
func SetLogLevel(level logrus.Level) {
	logger.Level = level
}

// SetLogFormatter ...
func SetLogFormatter(formatter logrus.Formatter) {
	logger.Formatter = formatter
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Debug(args...)
	}
}

// DebugWithFields Debug logs a message with fields at level Debug on the standard logger.
func DebugWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Debug(l)
	}
}

// Println Info logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	Info(context.Background(), args...)
}

// Info logs a message at level Info on the standard logger.
func Info(ctx context.Context, args ...interface{}) {
	fields := logrus.Fields{
		"service": os.Getenv("POD_CONTAINER"),
		"file":    fileInfo(2),
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		fields[TraceId] = traceID
	}

	if userId := GetUserId(ctx); userId != "" {
		fields[UserId] = userId
	}

	if logger.Level >= logrus.InfoLevel {
		logger.WithFields(fields).Info(args...)
	}
}

// InfoWithFields Debug logs a message with fields at level Debug on the standard logger.
func InfoWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Info(l)
	}
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Warn(args...)
	}
}

// WarnWithFields Debug logs a message with fields at level Debug on the standard logger.
func WarnWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Warn(l)
	}
}

func Error(ctx context.Context, args ...interface{}) {
	for _, v := range args {
		if err, ok := v.(error); ok {
			sentry.CaptureException(err)
		} else {
			sentry.CaptureMessage(fmt.Sprint(v))
		}
	}

	fields := logrus.Fields{
		"service": os.Getenv("POD_CONTAINER"),
		"file":    fileInfo(2),
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		fields[TraceId] = traceID
	}

	if userName := GetUserName(ctx); userName != "" {
		fields[UserName] = userName
	}

	if logger.Level >= logrus.ErrorLevel {
		logger.WithFields(fields).Error(args...)
	}
}

// ErrorWithFields Debug logs a message with fields at level Debug on the standard logger.
func ErrorWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Error(l)
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(args...)
	}
}

// FatalWithFields Debug logs a message with fields at level Debug on the standard logger.
func FatalWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(l)
	}
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	if logger.Level >= logrus.PanicLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Panic(args...)
	}
}

// PanicWithFields Debug logs a message with fields at level Debug on the standard logger.
func PanicWithFields(l interface{}, f Fields) {
	if logger.Level >= logrus.PanicLevel {
		entry := logger.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Panic(l)
	}
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func GetTraceID(ctx context.Context) string {
	traceId := metadata.ValueFromIncomingContext(ctx, TraceId)
	if len(traceId) == 0 {
		return ""
	}
	return traceId[0]
}

func GetUserId(ctx context.Context) string {
	userId := metadata.ValueFromIncomingContext(ctx, UserId)
	if len(userId) == 0 {
		return ""
	}
	return userId[0]
}

func GetUserName(ctx context.Context) string {
	userName := metadata.ValueFromIncomingContext(ctx, UserName)
	if len(userName) == 0 {
		return ""
	}
	return userName[0]
}

func AddContextFields(ctx context.Context, flds logrus.Fields) logrus.Fields {
	if flds == nil {
		flds = logrus.Fields{}
	}

	traceID := GetTraceID(ctx)

	if traceID != "" {
		flds[TraceId] = traceID
	}

	return flds
}
