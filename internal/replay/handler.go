package replay

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/driftwatch/internal/eventlog"
)

// PrintHandler returns a Handler that writes a human-readable line for each
// replayed event to w. If w is nil, os.Stderr is used.
func PrintHandler(w io.Writer) Handler {
	if w == nil {
		w = os.Stderr
	}
	return func(result interface{ GetPath() string }) {
		fmt.Fprintf(w, "[replay] %s drift detected on %s\n",
			time.Now().UTC().Format(time.RFC3339), result.GetPath())
	}
}

// RunWithPrint is a convenience wrapper that creates a Replayer, runs it, and
// prints each event to w using PrintHandler.
func RunWithPrint(ctx context.Context, log *eventlog.EventLog, opts Options, w io.Writer) error {
	r := New(log, opts)
	return r.Run(ctx, func(res interface{ GetPath() string }) {
		PrintHandler(w)(res)
	})
}
