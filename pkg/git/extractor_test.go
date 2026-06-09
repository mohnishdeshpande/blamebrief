package git

import (
	"testing"
)

func TestParseRawLog(t *testing.T) {
	rawLogInput := `commit 1aa4ec35e1f88da24ddf3fd38ef7449b3fe98da2
Author: Hackathon Archaeologist <archaeologist@deepmind-hack.internal>
Date:   Sat Jun 6 12:38:27 2026 +1000

    feat: Initial commit of BlameBrief CLI core structure

diff --git a/main.go b/main.go
index 0000000..8af14fe
--- a/main.go
+++ b/main.go
@@ -0,0 +1,10 @@
+package main
+
+import "fmt"
`

	commits := ParseRawLog(rawLogInput)

	if len(commits) != 1 {
		t.Fatalf("expected 1 parsed commit, got %d", len(commits))
	}

	c := commits[0]
	if c.Hash != "1aa4ec35e1f88da24ddf3fd38ef7449b3fe98da2" {
		t.Errorf("expected hash to be '1aa4ec35e1f88da24ddf3fd38ef7449b3fe98da2', got %q", c.Hash)
	}

	if c.Author != "Hackathon Archaeologist <archaeologist@deepmind-hack.internal>" {
		t.Errorf("expected author to match, got %q", c.Author)
	}

	if c.Date != "Sat Jun 6 12:38:27 2026 +1000" {
		t.Errorf("expected date to match, got %q", c.Date)
	}

	expectedMessage := "feat: Initial commit of BlameBrief CLI core structure"
	if c.Message != expectedMessage {
		t.Errorf("expected message to be %q, got %q", expectedMessage, c.Message)
	}

	if c.Diff == "" {
		t.Errorf("expected non-empty diff string")
	}
}
