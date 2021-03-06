package cloudlog

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
)

// ScopedLogger log information to Stackdrive console to be grouped based on the request
type ScopedLogger struct {
	entryLogger   *logging.Logger
	parentLogger  *logging.Logger
	request       *http.Request
	logSeverities []logging.Severity
	startTime     time.Time
	endTime       time.Time
	local         bool
}

// NewScopedLogger constructs and returns a new ScopedLogger.
func NewScopedLogger(client *logging.Client, r *http.Request, name string) *ScopedLogger {
	const (
		// parentFormat is a format string for a ScopedLogger's parent log name.
		parentFormat = "%v-request"
		// childFormat is a format string for a ScopedLogger's child log name.
		childFormat = "%v-entry"
	)
	// To aggregate all logs under the same resource tab
	customResource := &mrpb.MonitoredResource{
		Type: "gce_instance",
	}
	parentLogger := client.Logger(
		fmt.Sprintf(parentFormat, name),
		logging.CommonResource(customResource),
		logging.CommonLabels(WithHostname(nil)),
	)
	childLogger := client.Logger(
		fmt.Sprintf(childFormat, name),
		logging.CommonResource(customResource),
		logging.CommonLabels(WithHostname(nil)),
	)
	startTime := time.Now()
	endTime := startTime
	return &ScopedLogger{
		entryLogger:   childLogger,
		parentLogger:  parentLogger,
		request:       r,
		logSeverities: nil,
		startTime:     startTime,
		endTime:       endTime,
		local:         false,
	}
}

func (l *ScopedLogger) EnableLocal(flag bool) {
	l.local = flag
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
		HTTPRequest: &logging.HTTPRequest{
			Request: l.request,
		},
		Payload:  payload,
		Severity: severity,
	}
	l.entryLogger.Log(e)
	l.logSeverities = append(l.logSeverities, severity)
	if l.local {
		log.Printf("%v: %v", severity.String(), payload)
	}
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
			Latency: l.endTime.Sub(l.startTime),
			//Status:  200,
		},
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
		Severity: l.maxSeverity(),
		HTTPRequest: &logging.HTTPRequest{
			Request: l.request,
		},
	}

	l.parentLogger.Log(e)
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
