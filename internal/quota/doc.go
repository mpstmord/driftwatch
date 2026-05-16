// Package quota provides a per-path alert-rate quota enforced over a rolling
// time window.
//
// A Quota is created with a maximum alert count and a window duration:
//
//	q := quota.New(5, time.Minute)
//
// Each call to Allow increments the counter for the given path. Once the
// counter reaches the maximum the path is suppressed for the remainder of the
// window. The counter resets automatically when the window expires.
//
// Middleware integrates the Quota into the drift-result processing pipeline:
//
//	mw := quota.NewMiddleware(q, nextHandler)
//	mw.Handle(ctx, result)
//
// Quota is safe for concurrent use.
package quota
