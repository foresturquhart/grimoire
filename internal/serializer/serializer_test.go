package serializer

import (
	"path/filepath"
	"testing"

	"github.com/foresturquhart/grimoire/internal/secrets"
)

func TestGetFindingsForFile(t *testing.T) {
	base := filepath.Join("/tmp", "base")
	info := &RedactionInfo{
		Enabled: true,
		BaseDir: base,
		Findings: []secrets.Finding{
			{Description: "a", Secret: "AAA", File: filepath.Join(base, "a.txt")},
			{Description: "b", Secret: "BBB", File: filepath.Join(base, "b.txt")},
		},
	}

	res := GetFindingsForFile(info, "a.txt", base)
	if len(res) != 1 || res[0].Secret != "AAA" {
		t.Fatalf("unexpected findings: %#v", res)
	}

	if out := GetFindingsForFile(info, "c.txt", base); len(out) != 0 {
		t.Fatalf("expected no findings, got %#v", out)
	}

	if GetFindingsForFile(nil, "a.txt", base) != nil {
		t.Fatalf("expected nil when info is nil")
	}
}

func TestRedactSecrets(t *testing.T) {
	content := "line1\nkey=SECRET\npassword=PASS\n"
	findings := []secrets.Finding{
		{Description: "key", Secret: "SECRET", Line: 2},
		{Description: "pass", Secret: "PASS", Line: 3},
	}
	got := RedactSecrets(content, findings)
	expected := "line1\nkey=[REDACTED SECRET: key]\npassword=[REDACTED SECRET: pass]\n"
	if got != expected {
		t.Errorf("redacted output\n%q\nwant\n%q", got, expected)
	}

	// general replacement without line numbers
	content2 := "token=ABCDEF"
	findings2 := []secrets.Finding{{Description: "tok", Secret: "ABCDEF"}}
	exp2 := "token=[REDACTED SECRET: tok]"
	if out := RedactSecrets(content2, findings2); out != exp2 {
		t.Errorf("redact general %q want %q", out, exp2)
	}
}
