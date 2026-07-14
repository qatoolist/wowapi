//go:build dev || test

package appmodel

const IsDevBuild = true

// handlePostSealMutation handles post-seal mutation attempts.
// In development or test builds (with dev or test build tags), it panics.
func handlePostSealMutation(err error) error {
	panic(err)
}
