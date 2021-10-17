package log

import (
	"os"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

type Fields map[string]interface{}

func (f Fields) F(k string, v interface{}) Fields {
	f[k] = v
	return f
}

func (f Fields) E(err error) Fields {
	return f.F("error", err)
}

func F(k string, v interface{}) Fields {
	return Fields{}.F(k, v)
}

func E(err error) Fields {
	return Fields{}.E(err)
}

var logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}).With().Logger().Level(zerolog.InfoLevel)

func DefaultLevel() string {
	return logger.GetLevel().String()
}

func SetLevel(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		Fatal("error parse log level", E(err))
	}
	logger = logger.Level(lvl)
}

func Fatal(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Fatal(), msg, fields...)
}

func Debug(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Debug(), msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Info(), msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Warn(), msg, fields...)
}

func Error(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Error(), msg, fields...)
}

func Panic(msg string, fields ...map[string]interface{}) {
	sendEvent(logger.Panic(), msg, fields...)
}

func sendEvent(event *zerolog.Event, msg string, fields ...map[string]interface{}) {
	if event == nil {
		return
	}
	if len(fields) > 0 {
		event.Fields(fields[0])
	}

	event.Msg(msg)
}
