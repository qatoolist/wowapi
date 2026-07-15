//go:build !dev && !test

package appmodel

const IsDevBuild = false

// handlePostSealMutation handles post-seal mutation attempts.
// In production (without dev or test build tags), it returns the error.
func handlePostSealMutation(err error) error {
	return err
}
