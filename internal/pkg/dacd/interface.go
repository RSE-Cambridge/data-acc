package dacd

type BrickManager interface {
	// Get the current hostname key that is being kept alive
	Hostname() string

	// Tidy up from previous shutdowns
	// , then start waiting for session actions
	// notify dacctl we are listening via keep alive key
	Startup()
	// Wait for any events to complete
	// then do any tidy up required for a graceful shutdown
	Shutdown()
}
