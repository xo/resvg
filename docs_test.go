package resvg_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// These tests read the repository and not the package. They hold the agent
// setup that D15 adopted from dbmeta D110, and the decision log in
// docs/PLAN.md.

// TestClaudeImportsAgents holds D15. AGENTS.md holds the rules, because Codex
// and the other agents read it, and CLAUDE.md imports it, so that Claude Code
// reads the same rules. A symbolic link is refused for the reason in
// TestSkillsAreCopies.
func TestClaudeImportsAgents(t *testing.T) {
	t.Parallel()
	fi, err := os.Lstat("CLAUDE.md")
	switch {
	case err != nil:
		t.Fatal(err)
	case !fi.Mode().IsRegular():
		t.Fatalf("CLAUDE.md is not an ordinary file, its mode is %v. Make it a file that holds @AGENTS.md. See D15", fi.Mode())
	}
	if got := strings.TrimSpace(read(t, "CLAUDE.md")); got != "@AGENTS.md" {
		t.Errorf("CLAUDE.md holds %q. It holds only @AGENTS.md, and the rules go in AGENTS.md. See D15", got)
	}
	if _, err := os.Stat("AGENTS.md"); err != nil {
		t.Errorf("CLAUDE.md imports AGENTS.md, which is missing: %v", err)
	}
}

// TestTheRootHoldsFourDocuments holds D15. A document in the root that is not
// one of the four is a document that nobody filed.
func TestTheRootHoldsFourDocuments(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"README.md": true, "AGENTS.md": true, "CLAUDE.md": true, "CONTRIBUTING.md": true,
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		if !allowed[e.Name()] {
			t.Errorf("%s is in the repository root. Only README, AGENTS, CLAUDE and CONTRIBUTING belong there, "+
				"and every other document goes in docs/. See D15", e.Name())
		}
	}
	for name := range allowed {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s in the repository root", name)
		}
	}
}

// TestEveryDocumentIsInBothTables makes sure that each document in docs/ is named
// in the table at the top of AGENTS.md and in the list in README.md. A
// document that nobody can find is a document that nobody reads.
func TestEveryDocumentIsInBothTables(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir("docs")
	if err != nil {
		t.Fatal(err)
	}
	tables := map[string]string{
		"AGENTS.md": read(t, "AGENTS.md"),
		"README.md": read(t, "README.md"),
	}
	var found int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		found++
		for name, body := range tables {
			if !strings.Contains(body, "docs/"+e.Name()) {
				t.Errorf("%s does not name docs/%s. Add it", name, e.Name())
			}
		}
	}
	if found == 0 {
		t.Error("docs/ holds no document, so this test tests nothing")
	}
}

// markdownLink matches a relative link in a document, and leaves a URL alone.
var markdownLink = regexp.MustCompile(`\]\((?:\./)?([^)#:\s]+)(?:#[^)]*)?\)`)

// TestEveryMarkdownLinkResolves makes sure that each relative link in a document
// names a file or a folder that exists.
func TestEveryMarkdownLinkResolves(t *testing.T) {
	t.Parallel()
	for _, path := range repoFiles(t, ".md") {
		for _, m := range markdownLink.FindAllStringSubmatch(read(t, path), -1) {
			target := filepath.Join(filepath.Dir(path), m[1])
			if _, err := os.Stat(target); err != nil {
				t.Errorf("%s: the link to %s does not resolve to %s", path, m[1], target)
			}
		}
	}
}

// decision is one decision in docs/PLAN.md.
type decision struct {
	num    int
	title  string
	status string
	line   string
}

// decisionHeading matches the heading of a decision, which is its number, its
// title and its status:
//
//	### D8. Each platform archive is a Go module of its own. Decided, amended by D9 and D13.
//
// The title is greedy, so that a title can hold a period, as in version.txt.
var decisionHeading = regexp.MustCompile(`(?m)^### D([1-9][0-9]*)\. (.+)\. ((?:Decided|Proposed)(?:, [^.]+)?)\.$`)

// decisions returns the decisions in docs/PLAN.md, in the order of the file.
func decisions(t *testing.T) []decision {
	t.Helper()
	plan := read(t, filepath.Join("docs", "PLAN.md"))
	var out []decision
	for _, m := range decisionHeading.FindAllStringSubmatch(plan, -1) {
		num, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, decision{num: num, title: m[2], status: m[3], line: strings.TrimPrefix(m[0], "### ")})
	}
	if all := regexp.MustCompile(`(?m)^### D[0-9]+.*$`).FindAllString(plan, -1); len(all) != len(out) {
		for _, h := range all {
			if !decisionHeading.MatchString(h) {
				t.Errorf("docs/PLAN.md: %q is not in the form \"### D<n>. <Title>. <Status>.\", "+
					"where the status is Decided or Proposed, and can add \", amends D<n>\"", h)
			}
		}
	}
	if len(out) == 0 {
		t.Fatal("docs/PLAN.md holds no decision")
	}
	return out
}

// TestThePlanIndexIsComplete compares the index at the top of docs/PLAN.md
// against the decision headings. A reader finds a decision by its row, so a
// missing row, a stale title or a stale status hides it. The row links to the
// anchor that GitHub makes from the heading, so a changed status changes the
// link too, and the error prints the row to write.
func TestThePlanIndexIsComplete(t *testing.T) {
	t.Parallel()
	plan := read(t, filepath.Join("docs", "PLAN.md"))
	rows := make(map[int]string)
	for _, m := range regexp.MustCompile(`(?m)^\| \[D([0-9]+)\]\(.*$`).FindAllStringSubmatch(plan, -1) {
		num, _ := strconv.Atoi(m[1])
		rows[num] = m[0]
	}
	written := make(map[int]bool)
	for i, d := range decisions(t) {
		if d.num != i+1 {
			t.Errorf("D%d is decision %d in docs/PLAN.md. Number the decisions from D1, in order, with no gap", d.num, i+1)
		}
		written[d.num] = true
		want := fmt.Sprintf("| [D%d](#%s) | %s | %s |", d.num, anchor(d.line), d.title, d.status)
		switch got, ok := rows[d.num]; {
		case !ok:
			t.Errorf("D%d has no row in the index of docs/PLAN.md. Add:\n%s", d.num, want)
		case got != want:
			t.Errorf("D%d: the row in the index of docs/PLAN.md is\n%s\nand the heading needs\n%s", d.num, got, want)
		}
	}
	for num := range rows {
		if !written[num] {
			t.Errorf("the index of docs/PLAN.md has a row for D%d, and no heading holds it", num)
		}
	}
}

// anchor returns the anchor that GitHub makes from the text of a heading: the
// text in lower case, with each character that is not a letter, a digit, a
// space, a hyphen or an underscore removed, and each space changed to a
// hyphen.
func anchor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// amendRef matches the part of a status that names other decisions, such as
// "amends D8" or "amended by D9 and D13".
var amendRef = regexp.MustCompile(`\b(amends|amended by) (D[0-9]+(?:(?:, | and )D[0-9]+)*)`)

// TestAnAmendmentPointsBothWays makes sure that when D9 says "amends D8", D8 says
// "amended by D9", and the other way round. A reader who finds the older
// decision has to learn that it no longer holds as written.
func TestAnAmendmentPointsBothWays(t *testing.T) {
	t.Parallel()
	status := make(map[int]string)
	for _, d := range decisions(t) {
		status[d.num] = d.status
	}
	opposite := map[string]string{"amends": "amended by", "amended by": "amends"}
	var found int
	for num, s := range status {
		for verb, others := range amendments(s) {
			for _, other := range others {
				found++
				if _, ok := status[other]; !ok {
					t.Errorf("D%d says %q, and D%d is not a decision", num, s, other)
					continue
				}
				if !slices.Contains(amendments(status[other])[opposite[verb]], num) {
					t.Errorf("D%d says %q, and D%d does not say \"%s D%d\". "+
						"An amendment is written in the status of both decisions", num, s, other, opposite[verb], num)
				}
			}
		}
	}
	if found == 0 {
		t.Error("no decision amends another, so this test tests nothing")
	}
}

// amendments returns the decisions that a status names, keyed by "amends" or
// "amended by".
func amendments(status string) map[string][]int {
	out := make(map[string][]int)
	for _, m := range amendRef.FindAllStringSubmatch(status, -1) {
		for _, ref := range regexp.MustCompile(`D([0-9]+)`).FindAllStringSubmatch(m[2], -1) {
			num, _ := strconv.Atoi(ref[1])
			out[m[1]] = append(out[m[1]], num)
		}
	}
	return out
}

// decisionRef matches a reference to a decision by its number, such as D3.
var decisionRef = regexp.MustCompile(`\bD([1-9][0-9]*)\b`)

// repositories are the names that can come before a decision of another
// repository, as in "dbmeta D110".
var repositories = map[string]bool{
	"blitz": true, "cql": true, "dbimp": true, "dbmeta": true, "dbtpl": true,
	"dburl": true, "tblfmt": true, "usql": true, "vegagoja": true,
}

// TestEveryDecisionReferenceExists makes sure that a bare decision number in a
// document or a Go file is a decision in docs/PLAN.md. A reference to a
// decision of another repository names that repository first, as in
// "dbmeta D110", and this test skips it.
func TestEveryDecisionReferenceExists(t *testing.T) {
	t.Parallel()
	written := make(map[string]bool)
	for _, d := range decisions(t) {
		written[strconv.Itoa(d.num)] = true
	}
	for _, path := range repoFiles(t, ".md", ".go") {
		body := read(t, path)
		for _, m := range decisionRef.FindAllStringSubmatchIndex(body, -1) {
			num := body[m[2]:m[3]]
			if written[num] {
				continue
			}
			before := strings.Fields(body[max(0, m[0]-40):m[0]])
			if len(before) != 0 && repositories[strings.Trim(before[len(before)-1], "(`\"")] {
				continue
			}
			t.Errorf("%s: names D%s, which is not a decision in docs/PLAN.md. "+
				"A decision of another repository names that repository, as in dbmeta D110", path, num)
		}
	}
}

// repoFiles returns each file with one of the given extensions. It skips the
// folders whose names start with a period, such as .git and the agent skills,
// which this repository does not write.
func repoFiles(t *testing.T, exts ...string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && path != "." && strings.HasPrefix(d.Name(), "."):
			return filepath.SkipDir
		case d.IsDir():
			return nil
		}
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				out = append(out, path)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
