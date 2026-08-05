package clitest

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Update is the shared -update flag: run `go test ./cmd/... -update` to
// rewrite every golden fixture from the current output instead of comparing
// against it. It is registered once here, in the package every cmd/*_test.go
// already imports for its child harness, rather than once per command.
//
// Eyeball a regenerated fixture's diff before committing it — -update writes
// whatever the command currently emits, bugs included.
var Update = flag.Bool("update", false, "update golden CLI-output fixtures instead of comparing against them")

// AssertGolden compares stdout against the fixture at testdata/<name>.txt,
// byte for byte. With -update it (re)writes the fixture from stdout instead.
//
// This is stdout only — clitest's AssertRejected/AssertReportedSeed already
// cover the exit-code/stderr half of the CLI contract; a golden fixture's job
// is catching a change in what a command actually RENDERS (a reordered
// field, a shifted dice stream moving every downstream value) that those
// checks cannot see, closing the #321 class of hazard structurally instead
// of relying on a person to notice a stale sample.
func (r Result) AssertGolden(t *testing.T, name string) {
	t.Helper()

	path := filepath.Join("testdata", name+".txt")

	if *Update {
		// 0o644, matching every other tracked text file: 0o600 would make a
		// freshly-created fixture unreadable to anyone but its owner, and
		// would disagree with core.filemode's usual 644 the moment a second
		// fixture in the same directory needs group/other read access.
		err := os.WriteFile(path, []byte(r.Stdout), 0o644) //nolint:gosec // golden fixture, not a secret
		if err != nil {
			t.Fatalf("writing golden %s: %v", path, err)
		}

		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden %s: %v (run `go test ./... -update` to create it)", path, err)
	}

	if r.Stdout != string(want) {
		t.Errorf("stdout does not match golden %s\n--- got ---\n%s--- want ---\n%s", path, r.Stdout, want)
	}
}

// ReadmeBlock reads the file at readmePath and returns the run of non-blank
// lines immediately following the line containing marker, joined with "\n"
// and given a trailing newline — the shape of a fmt.Println-based command's
// captured stdout. marker is matched as a substring so callers can pass the
// distinctive "$ go run ..." line from the README's fenced sample block
// without escaping the rest of it.
//
// It fails t if marker is not found, rather than returning an empty block
// that would make every comparison trivially pass.
func ReadmeBlock(t *testing.T, readmePath, marker string) string {
	t.Helper()

	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("reading %s: %v", readmePath, err)
	}

	// Normalize CRLF/CR to LF before splitting: a Windows checkout with
	// core.autocrlf can give README.md \r\n line endings, which would leave a
	// trailing \r on every captured line and fail every comparison against a
	// command's LF-only stdout even when the visible text is identical.
	normalized := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")

	lines := strings.Split(normalized, "\n")

	start := -1

	for i, line := range lines {
		if strings.Contains(line, marker) {
			start = i + 1

			break
		}
	}

	if start < 0 {
		t.Fatalf("marker %q not found in %s", marker, readmePath)
	}

	var block []string

	for _, line := range lines[start:] {
		// A blank line separates samples within one fenced block; a code fence
		// ends the last sample in it, which has no trailing blank line before it.
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "```") {
			break
		}

		block = append(block, line)
	}

	if len(block) == 0 {
		t.Fatalf("marker %q in %s is followed by no sample lines", marker, readmePath)
	}

	return strings.Join(block, "\n") + "\n"
}

// AssertReadmeUpToDate fails t if r.Stdout diverges from the README sample
// block starting after marker — the specific #321 failure (a stale pinned
// sample, caught only by luck in review) this issue exists to close
// structurally.
//
// Trailing newlines are trimmed from both sides before comparing: the
// README's blank line after a sample is markdown's own block-boundary
// marker, not literal command output, so it carries no information about
// how many trailing newlines the command itself prints (a sheet's closing
// blank line, say). AssertGolden already pins stdout byte for byte,
// including any such trailing whitespace; this check is about content
// drift, not trailing-newline count.
func (r Result) AssertReadmeUpToDate(t *testing.T, readmePath, marker string) {
	t.Helper()

	want := strings.TrimRight(ReadmeBlock(t, readmePath, marker), "\n")
	got := strings.TrimRight(r.Stdout, "\n")

	if got != want {
		t.Errorf("README sample is stale.\n--- README shows ---\n%s\n--- actual output ---\n%s", want, got)
	}
}
