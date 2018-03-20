package cloudlog

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
)

// Logger is the logger to send logging information to Stackdrive console for your Google cloud project
type Logger struct {
	logger *logging.Logger
}

// Configure configures an returns the Stackdrive logger. Caller is responsible to guarantee the right permission to initialize the logging client
func Configure(projectID string, loggerID string) (*Logger, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Logger{logger: client.Logger(loggerID)}, nil
}

func (l *Logger) output(payload string, severity logging.Severity) {
	e := logging.Entry{
		Payload:  payload,
		Severity: severity,
	}
	l.logger.Log(e)
}

// Debug logs the payload
func (l *Logger) Debug(payload string) {
	l.output(payload, logging.Debug)
}

// Debugf formats according to a format specifier and logs it
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Debug(fmt.Sprintf(format, v))
}

// Info logs the payload
func (l *Logger) Info(payload string) {
	l.output(payload, logging.Info)
}

// Infof formats according to a format specifier and logs it
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v))
}

// Warning logs the payload
func (l *Logger) Warning(payload string) {
	l.output(payload, logging.Warning)
}

// Warningf formats according to a format specifier and logs it
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Warning(fmt.Sprintf(format, v))
}

// Error logs the payload
func (l *Logger) Error(payload string) {
	l.output(payload, logging.Error)
}

// Errorf formats according to a format specifier and logs it
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v))
}

// Critical logs the payload
func (l *Logger) Critical(payload string) {
	l.output(payload, logging.Critical)
}

// Criticalf formats according to a format specifier and logs it
func (l *Logger) Criticalf(format string, v ...interface{}) {
	l.Critical(fmt.Sprintf(format, v))
}

// Alert logs the payload
func (l *Logger) Alert(payload string) {
	l.output(payload, logging.Alert)
}

// Alertf formats according to a format specifier and logs it
func (l *Logger) Alertf(format string, v ...interface{}) {
	l.Alert(fmt.Sprintf(format, v))
}

// Emergency logs the payload
func (l *Logger) Emergency(payload string) {
	l.output(payload, logging.Emergency)
}

// Emergencyf formats according to a format specifier and logs it
func (l *Logger) Emergencyf(format string, v ...interface{}) {
	l.Emergency(fmt.Sprintf(format, v))
}
