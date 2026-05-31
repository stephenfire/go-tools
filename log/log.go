package log

import (
	"errors"
	"os"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var (
	rootLog    *logrus.Logger
	wrapped    logrus.FieldLogger
	_logLocker sync.RWMutex
)

type EmptyWriter struct{}

func (w *EmptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func InitLogLevelAndFile(logLevel string, logPath string, logPathSuffix string) {
	level := logrus.DebugLevel
	if logLevel != "" {
		if ll, err := logrus.ParseLevel(logLevel); err != nil {
			Warnf("invalid log level %s, using default %s", logLevel, level)
		} else {
			level = ll
		}
	}
	if logPath != "" {
		Infof("log path:%s, level:%s", logPath, level)
		InitLogWithSuffix(logPath, logPathSuffix, level)
	} else {
		Infof("log level:%s", level)
		_ = SetLevel(level)
	}
}

func InitLogWithSuffix(path string, suffix string, level logrus.Level) {
	_logLocker.Lock()
	defer _logLocker.Unlock()
	writer, err := rotatelogs.New(
		path+"."+suffix+".%Y%m%d",
		rotatelogs.WithMaxAge(time.Duration(86400*30)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
	)
	if err != nil {
		panic("failed to create rotatelogs: " + path)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: time.StampMilli,
	}

	rootLog = &logrus.Logger{
		Out:       &EmptyWriter{},
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}
	rootLog.AddHook(NewFileAndConsoleHook(formatter, writer, os.Stdout,
		logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel))
	wrapped = rootLog
}

func SetLevel(level logrus.Level) error {
	switch level {
	case logrus.TraceLevel, logrus.DebugLevel:
	case logrus.InfoLevel:
	case logrus.WarnLevel:
	case logrus.ErrorLevel:
	case logrus.PanicLevel:
	case logrus.FatalLevel:
	default:
		return errors.New("invalid log level")
	}
	rootLog.SetLevel(level)
	return nil
}

func init() {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	}

	rootLog = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.TraceLevel,
	}
	wrapped = rootLog
}

// func Logger() *logrus.Logger {
// 	return rootLog
// }

func Debug(msgs ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Debug(msgs...)
}

func Debugf(format string, values ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Debugf(format, values...)
}

func Info(msgs ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Info(msgs...)
}

func Infof(format string, values ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Infof(format, values...)
}

func Warn(msgs ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Warn(msgs...)
}

func Warnf(format string, values ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Warnf(format, values...)
}

func Error(msgs ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Error(msgs...)
}

func Errorf(format string, values ...interface{}) {
	_logLocker.RLock()
	defer _logLocker.RUnlock()

	wrapped.Errorf(format, values...)
}

func MustDebugf(logger logrus.FieldLogger, format string, args ...interface{}) {
	if logger == nil {
		Debugf(format, args...)
	} else {
		logger.Debugf(format, args...)
	}
}

func MustInfof(logger logrus.FieldLogger, format string, args ...interface{}) {
	if logger == nil {
		Infof(format, args...)
	} else {
		logger.Infof(format, args...)
	}
}

func MustWarnf(logger logrus.FieldLogger, format string, args ...interface{}) {
	if logger == nil {
		Warnf(format, args...)
	} else {
		logger.Warnf(format, args...)
	}
}

func MustErrorf(logger logrus.FieldLogger, format string, args ...interface{}) {
	if logger == nil {
		Errorf(format, args...)
	} else {
		logger.Errorf(format, args...)
	}
}

func WithFields(fields logrus.Fields) logrus.FieldLogger {
	return rootLog.WithFields(fields)
}

func WithField(vs ...interface{}) logrus.FieldLogger {
	if len(vs) <= 1 {
		return rootLog
	}
	l := len(vs)
	fields := make(logrus.Fields)
	for i := 0; i < l/2; i += 2 {
		k, ok := vs[i].(string)
		if !ok {
			continue
		}
		fields[k] = vs[i+1]
	}
	return rootLog.WithFields(fields)
}

func SetFields(fields logrus.Fields) {
	_logLocker.Lock()
	defer _logLocker.Unlock()

	logger := wrapped.WithFields(fields)
	wrapped = logger
}
