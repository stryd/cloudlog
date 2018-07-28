package cloudlog

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
)

// ScopedLogger log information to Stackdrive console to be grouped based on the request
type ScopedLogger struct {
	entryLogger   *logging.Logger
	parentLogger  *logging.Logger
	request       *http.Request
	logSeverities []logging.Severity
	traceID       string
	startTime     time.Time
	endTime       time.Time
}

// NewScopedLogger constructs and returns a new ScopedLogger.
func NewScopedLogger(client *logging.Client, r *http.Request, name string) *ScopedLogger {
	const (
		// parentFormat is a format string for a ScopedLogger's parent log name.
		parentFormat = "%v-request"
		// childFormat is a format string for a ScopedLogger's child log name.
		childFormat = "%v-entry"
	)
	parentLogger := client.Logger(
		fmt.Sprintf(parentFormat, name),
		logging.CommonLabels(WithHostname(nil)),
	)
	childLogger := client.Logger(
		fmt.Sprintf(childFormat, name),
		logging.CommonLabels(WithHostname(nil)),
	)
	startTime := time.Now()
	endTime := startTime
	return &ScopedLogger{
		entryLogger:   childLogger,
		parentLogger:  parentLogger,
		request:       r,
		logSeverities: nil,
		traceID:       getTraceID(r),
		startTime:     startTime,
		endTime:       endTime,
	}
}

func (l *ScopedLogger) maxSeverity() logging.Severity {
	maxSeverity := logging.Default
	for _, s := range l.logSeverities {
		if s > maxSeverity {
			maxSeverity = s
		}
	}
	return maxSeverity
}

func (l *ScopedLogger) output(payload string, severity logging.Severity) {
	e := logging.Entry{
		Payload:  payload,
		Severity: severity,
		Trace:    l.traceID,
	}
	l.entryLogger.Log(e)
	l.logSeverities = append(l.logSeverities, severity)
}

// Debug logs the payload
func (l *ScopedLogger) Debug(payload string) {
	l.output(payload, logging.Debug)
}

// Debugf formats according to a format specifier and logs it
func (l *ScopedLogger) Debugf(format string, v ...interface{}) {
	l.Debug(fmt.Sprintf(format, v...))
}

// Info logs the payload
func (l *ScopedLogger) Info(payload string) {
	l.output(payload, logging.Info)
}

// Infof formats according to a format specifier and logs it
func (l *ScopedLogger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

// Warning logs the payload
func (l *ScopedLogger) Warning(payload string) {
	l.output(payload, logging.Warning)
}

// Warningf formats according to a format specifier and logs it
func (l *ScopedLogger) Warningf(format string, v ...interface{}) {
	l.Warning(fmt.Sprintf(format, v...))
}

// Error logs the payload
func (l *ScopedLogger) Error(payload string) {
	l.output(payload, logging.Error)
}

// Errorf formats according to a format specifier and logs it
func (l *ScopedLogger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

// Critical logs the payload
func (l *ScopedLogger) Critical(payload string) {
	l.output(payload, logging.Critical)
}

// Criticalf formats according to a format specifier and logs it
func (l *ScopedLogger) Criticalf(format string, v ...interface{}) {
	l.Critical(fmt.Sprintf(format, v...))
}

// Alert logs the payload
func (l *ScopedLogger) Alert(payload string) {
	l.output(payload, logging.Alert)
}

// Alertf formats according to a format specifier and logs it
func (l *ScopedLogger) Alertf(format string, v ...interface{}) {
	l.Alert(fmt.Sprintf(format, v...))
}

// Emergency logs the payload
func (l *ScopedLogger) Emergency(payload string) {
	l.output(payload, logging.Emergency)
}

// Emergencyf formats according to a format specifier and logs it
func (l *ScopedLogger) Emergencyf(format string, v ...interface{}) {
	l.Emergency(fmt.Sprintf(format, v...))
}

// Finish doesn't log any payload, it just provides the http request, response size and status code
func (l *ScopedLogger) Finish() {
	l.endTime = time.Now()
	e := logging.Entry{
		HTTPRequest: &logging.HTTPRequest{
			Request: l.request,
			//Status:  statusCode,
			Latency: l.endTime.Sub(l.startTime),
		},
		/*
			Operation: &logpb.LogEntryOperation{
				Id:       appID,
				Producer: "backend.stryd.com",
			},
		*/
		Trace:    l.traceID,
		Severity: l.maxSeverity(),
	}
	l.parentLogger.Log(e)
	l.parentLogger.Flush()
}

// partialFinish is called after N seconds to provide a log entry in the
// main Stackdriver log stream to aggregate the child logs. Should only be
// called for long running requests
func (l *ScopedLogger) partialFinish() {
	e := logging.Entry{
		Trace:    l.traceID,
		Severity: l.maxSeverity(),
		HTTPRequest: &logging.HTTPRequest{
			Request: l.request,
		},
	}

	l.parentLogger.Log(e)
}

// getTraceID is an ID by which the group will be grouped in the Google
// Cloud Logging console.
//
// If the `X-Cloud-Trace-Context` header is set in the request by GCP
// middleware, then that trace ID is used.
//
// Otherwise, a pseudorandom UUID is used.
func getTraceID(r *http.Request) string {
	// If the trace header exists, use the trace.
	if id := r.Header.Get("X-Cloud-Trace-Context"); id != "" {
		return id
	}
	// Otherwise, generate a random group ID.
	return uuid.New().String()
}

var detectedHost struct {
	hostname string
	once     sync.Once
}

// WithHostname adds the hostname to a labels map. Useful for setting common
// labels: logging.CommonLabels(WithHostname(labels))
func WithHostname(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	detectedHost.once.Do(func() {
		if metadata.OnGCE() {
			instanceName, err := metadata.InstanceName()
			if err == nil {
				detectedHost.hostname = instanceName
			}
		} else {
			hostname, err := os.Hostname()
			if err == nil {
				detectedHost.hostname = hostname
			}
		}
	})
	labels["hostname"] = detectedHost.hostname
	return labels
}
