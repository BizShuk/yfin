// exitcodes.go — process exit code constants shared by every cobra subcommand
// in cmd/. The codes mirror the yfin CLI requirements: 0 success; 1 generic
// failure; 2 paid-only endpoint; 3 user/configuration error.
package cmd

// Exit codes as specified in the requirements
const (
	ExitSuccess     = 0
	ExitGeneral     = 1
	ExitPaidFeature = 2
	ExitConfigError = 3
)
