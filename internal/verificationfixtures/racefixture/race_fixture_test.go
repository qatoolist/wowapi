//go:build wowapi_race_fixture

package racefixture

import (
	"sync"
	"testing"
)

// TestSeededDataRace is deliberately excluded from normal builds. The
// check_race_detector.sh negative fixture enables it and requires -race to
// diagnose the unsynchronized writes before the real integration race suite.
func TestSeededDataRace(t *testing.T) {
	var value int
	start := make(chan struct{})
	var writers sync.WaitGroup
	writers.Add(2)
	for range 2 {
		go func() {
			defer writers.Done()
			<-start
			for range 10_000 {
				value++
			}
		}()
	}
	close(start)
	writers.Wait()
	_ = value
}
