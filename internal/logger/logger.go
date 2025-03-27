package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Level = logrus.Level

var (
	version     = ""
	serviceName = os.Getenv("SERVICE_NAME")
)

var (
	defaultFormatter *logrus.TextFormatter
	log              *Logger
)

func init() {
	defaultFormatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	log = New(os.Stdout, logrus.InfoLevel)
	log.Logger.Formatter = defaultFormatter

	log.Logger.AddHook(LogHook{})
}

type LogHook struct{}

func (h LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h LogHook) Fire(entry *logrus.Entry) error {
	if serviceName != "" {
		entry.Data["file"] = serviceName
	}
	if version != "" {
		entry.Data["version"] = version
	}
	return nil
}

type Logger struct {
	*logrus.Logger
	level Level
	out   io.Writer
}

func New(out io.Writer, level Level) *Logger {
	logger := logrus.New()
	return &Logger{
		Logger: logger,
		level:  level,
		out:    out,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.out = w
	l.Logger.Out = l.out
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
	l.Logger.SetLevel(logrus.Level(l.level))
}

type callInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

func retrieveCallInfo(skip int) *callInfo {
	pc, file, line, _ := runtime.Caller(skip)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return &callInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

// Log messages with package name and filename
func (l *Logger) Log(level logrus.Level, msgf string, args ...any) {
	ci := retrieveCallInfo(3)
	if l.Logger.GetLevel() <= logrus.Level(level) {
		var msg string
		if len(args) > 0 {
			msg = fmt.Sprintf(msgf, args...)
		} else {
			msg = msgf
		}
		l.Logf(level, "%s(%s:%d): %s", filepath.Base(os.Args[0]), ci.fileName, ci.line, msg)
	}
}

func Debug(msg string) {
	log.Log(logrus.DebugLevel, msg)
}

func Info(msg string) {
	log.Log(logrus.InfoLevel, msg)
}

func Warn(msg string) {
	log.Log(logrus.WarnLevel, msg)
}

func Error(msg string) {
	log.Log(logrus.ErrorLevel, msg)
}

func Debugf(msg string, args ...any) {
	log.Log(logrus.DebugLevel, msg, args...)
}

func Infof(msg string, args ...any) {
	log.Log(logrus.InfoLevel, msg, args...)
}

func Warnf(msg string, args ...any) {
	log.Log(logrus.WarnLevel, msg, args...)
}

func Errorf(msg string, args ...any) {
	log.Log(logrus.ErrorLevel, msg, args...)
}

func SetVersion(s string) {
	version = s
}

func SetServiceName(s string) {
	serviceName = s
}
