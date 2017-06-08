package logger

import (
	"fmt"
	"os"
	"gommo/config"
	"github.com/Sirupsen/logrus"
	"encoding/json"
)

type simpleLogFormatter struct{
	Type			string
	TimestampFormat string
}

func (f *simpleLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/377
			fields[k] = v.Error()
		default:
			fields[k] = v
		}
	}

	fields["@version"] = "1"

	timeStampFormat := f.TimestampFormat

	if timeStampFormat == "" {
		timeStampFormat = logrus.DefaultTimestampFormat
	}

	fields["@timestamp"] = entry.Time.Format(timeStampFormat)

	// set message field
	v, ok := entry.Data["message"]
	if ok {
		fields["fields.message"] = v
	}
	fields["message"] = entry.Message

	// set level field
	v, ok = entry.Data["level"]
	if ok {
		fields["fields.level"] = v
	}
	fields["level"] = entry.Level.String()

	// set type field
	if f.Type != "" {
		v, ok = entry.Data["type"]
		if ok {
			fields["fields.type"] = v
		}
		fields["type"] = f.Type
	}

	serialized, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func InitLogger(conf config.LoggerConfig) (func(), error) {
	level, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		fmt.Printf("error parse log level: %s\n", err)
	} else {
		logrus.SetLevel(level)
	}

	logrus.SetFormatter(&simpleLogFormatter{})

	// если имя файла не указано, используем вывод на консоль
	filename := conf.LogFile
	if filename != "" {
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logrus.SetOutput(f)
		return func() { f.Close() }, nil
	}
	return func() {}, nil
}
