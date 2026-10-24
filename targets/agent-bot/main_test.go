package main

import (
	"testing"

	shim "embedded-skills/shim"
)

func TestCorpusIsValid(t *testing.T) {
	cat, err := shim.LoadCatalog(corpus)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Skills) == 0 {
		t.Fatal("empty corpus")
	}
	for _, s := range cat.Skills {
		if len(s.Description) > 200 {
			t.Errorf("%s: description %d chars > 200", s.Name, len(s.Description))
		}
		for _, link := range append(s.Prerequisites, s.Next...) {
			if cat.Find(link) == nil {
				t.Errorf("%s: dangling chain link %q", s.Name, link)
			}
		}
	}
	// Manifest checksums must match embedded bytes (staleness gate).
	for _, ms := range cat.Manifest.Skills {
		data, err := corpus.ReadFile(ms.Path)
		if err != nil {
			t.Fatal(err)
		}
		if got := shim.Checksum(data); got != ms.Checksum {
			t.Errorf("%s: checksum %s != manifest %s (run make manifest)", ms.Path, got, ms.Checksum)
		}
	}
}
