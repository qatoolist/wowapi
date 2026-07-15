package requestbench

import "testing"

func TestEachProfileHasARepresentativeQueryPlanHash(t *testing.T) {
	s := newRequestSuite(t)
	unique := map[string]struct{}{}
	for _, profile := range workloadProfiles {
		hash := s.planHashes[profile]
		if hash == "" {
			t.Fatalf("profile %q has no plan hash", profile)
		}
		unique[hash] = struct{}{}
	}
	if len(unique) < 4 {
		t.Fatalf("only %d distinct plan hashes for six different request profiles", len(unique))
	}
}
