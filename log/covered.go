package log

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type CoveredLogger struct {
	logger   logrus.FieldLogger
	coverMap map[logrus.Level]logrus.Level
}

func NewCoveredLogger(logger logrus.FieldLogger, covers ...logrus.Level) logrus.FieldLogger {
	if logger == nil {
		panic(errors.New("logger cannot be nil"))
	}
	if len(covers) == 0 {
		return logger
	}
	if len(covers)%2 != 0 {
		panic(errors.New("number of the covers must be even"))
	}
	coverMap := make(map[logrus.Level]logrus.Level, len(covers)/2)
	for i := 0; i < len(covers); i += 2 {
		coverMap[covers[i]] = covers[i+1]
	}
	return &CoveredLogger{
		logger:   logger,
		coverMap: coverMap,
	}
}

func (l *CoveredLogger) _f(level logrus.Level, format string, args ...interface{}) {
	lvl, ok := l.coverMap[level]
	if !ok {
		lvl = level
	}
	switch lvl {
	case logrus.TraceLevel:
		l.logger.Debugf(format, args...)
	case logrus.DebugLevel:
		l.logger.Debugf(format, args...)
	case logrus.InfoLevel:
		l.logger.Infof(format, args...)
	case logrus.WarnLevel:
		l.logger.Warnf(format, args...)
	case logrus.ErrorLevel:
		l.logger.Errorf(format, args...)
	case logrus.FatalLevel:
		l.logger.Fatalf(format, args...)
	case logrus.PanicLevel:
		l.logger.Panicf(format, args...)
	}
}

func (l *CoveredLogger) _l(level logrus.Level, args ...interface{}) {
	lvl, ok := l.coverMap[level]
	if !ok {
		lvl = level
	}
	switch lvl {
	case logrus.TraceLevel:
		l.logger.Debug(args...)
	case logrus.DebugLevel:
		l.logger.Debug(args...)
	case logrus.InfoLevel:
		l.logger.Info(args...)
	case logrus.WarnLevel:
		l.logger.Warn(args...)
	case logrus.ErrorLevel:
		l.logger.Error(args...)
	case logrus.FatalLevel:
		l.logger.Fatal(args...)
	case logrus.PanicLevel:
		l.logger.Panic(args...)
	}
}

func (l *CoveredLogger) _ln(level logrus.Level, args ...interface{}) {
	lvl, ok := l.coverMap[level]
	if !ok {
		lvl = level
	}
	switch lvl {
	case logrus.TraceLevel:
		l.logger.Debugln(args...)
	case logrus.DebugLevel:
		l.logger.Debugln(args...)
	case logrus.InfoLevel:
		l.logger.Infoln(args...)
	case logrus.WarnLevel:
		l.logger.Warnln(args...)
	case logrus.ErrorLevel:
		l.logger.Errorln(args...)
	case logrus.FatalLevel:
		l.logger.Fatalln(args...)
	case logrus.PanicLevel:
		l.logger.Panicln(args...)
	}
}

func (l *CoveredLogger) WithField(key string, value interface{}) *logrus.Entry {
	return l.logger.WithField(key, value)
}

func (l *CoveredLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.logger.WithFields(fields)
}

func (l *CoveredLogger) WithError(err error) *logrus.Entry {
	return l.logger.WithError(err)
}

func (l *CoveredLogger) Debugf(format string, args ...interface{}) {
	l._f(logrus.DebugLevel, format, args...)
}
func (l *CoveredLogger) Infof(format string, args ...interface{}) {
	l._f(logrus.InfoLevel, format, args...)
}
func (l *CoveredLogger) Printf(format string, args ...interface{}) {
	l._f(logrus.InfoLevel, format, args...)
}
func (l *CoveredLogger) Warnf(format string, args ...interface{}) {
	l._f(logrus.WarnLevel, format, args...)
}
func (l *CoveredLogger) Warningf(format string, args ...interface{}) {
	l._f(logrus.WarnLevel, format, args...)
}
func (l *CoveredLogger) Errorf(format string, args ...interface{}) {
	l._f(logrus.ErrorLevel, format, args...)
}
func (l *CoveredLogger) Fatalf(format string, args ...interface{}) {
	l._f(logrus.FatalLevel, format, args...)
}
func (l *CoveredLogger) Panicf(format string, args ...interface{}) {
	l._f(logrus.PanicLevel, format, args...)
}

func (l *CoveredLogger) Debug(args ...interface{})   { l._l(logrus.DebugLevel, args...) }
func (l *CoveredLogger) Info(args ...interface{})    { l._l(logrus.InfoLevel, args...) }
func (l *CoveredLogger) Print(args ...interface{})   { l._l(logrus.InfoLevel, args...) }
func (l *CoveredLogger) Warn(args ...interface{})    { l._l(logrus.WarnLevel, args...) }
func (l *CoveredLogger) Warning(args ...interface{}) { l._l(logrus.WarnLevel, args...) }
func (l *CoveredLogger) Error(args ...interface{})   { l._l(logrus.ErrorLevel, args...) }
func (l *CoveredLogger) Fatal(args ...interface{})   { l._l(logrus.FatalLevel, args...) }
func (l *CoveredLogger) Panic(args ...interface{})   { l._l(logrus.PanicLevel, args...) }

func (l *CoveredLogger) Debugln(args ...interface{})   { l._ln(logrus.DebugLevel, args...) }
func (l *CoveredLogger) Infoln(args ...interface{})    { l._ln(logrus.InfoLevel, args...) }
func (l *CoveredLogger) Println(args ...interface{})   { l._ln(logrus.InfoLevel, args...) }
func (l *CoveredLogger) Warnln(args ...interface{})    { l._ln(logrus.WarnLevel, args...) }
func (l *CoveredLogger) Warningln(args ...interface{}) { l._ln(logrus.WarnLevel, args...) }
func (l *CoveredLogger) Errorln(args ...interface{})   { l._ln(logrus.ErrorLevel, args...) }
func (l *CoveredLogger) Fatalln(args ...interface{})   { l._ln(logrus.FatalLevel, args...) }
func (l *CoveredLogger) Panicln(args ...interface{})   { l._ln(logrus.PanicLevel, args...) }
