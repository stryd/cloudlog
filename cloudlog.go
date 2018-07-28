// Logging library to enable request scoped logging feature on StackDriver
//
// Example (For ScopedLogger):
//
//      ctx := context.Background()
//		loggingClient, err := NewClient(ctx, "your-project-ID")
//		if err != nil {
//			// Handle "failed to generate Stackdriver client."
//		}
//
//		var r *http.Request
//		name := "logger-id"
//      logger := cloudlog.NewScopedLogger(loggingClient, r, name)
//
//		logger.Info("Info log entry body.")
//		logger.Error("Error log entry body.")
//
//      logger.Finish()	// If you want to have the scoped logs. Otherwise all the logs will appear as individual entry

package cloudlog

import (
	"context"

	"cloud.google.com/go/logging"
)

// Configure generates a new logging client associated with the provided project
// parent has to be the projectID if you want to see the logs in GCP interface.
func Configure(ctx context.Context, parent string) (*logging.Client, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return nil, err
	}
	return client, nil
}
