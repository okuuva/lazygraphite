package git_commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

// GraphiteEntry holds parsed gt ls output for a single branch.
type GraphiteEntry struct {
	Name       string
	Prefix     string // tree-drawing prefix from gt ls
	IsCurrent  bool   // ◉ or ● in gt ls
	IsTrunk    bool
	StackIndex int // color index based on column position of circle
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// LoadGraphiteEntries runs `gt ls` and parses each line into a branch name,
// tree prefix, and column-based color index. Returns nil if not a Graphite repo.
func LoadGraphiteEntries(repoGitDir string) []GraphiteEntry {
	dbPath := filepath.Join(repoGitDir, ".graphite_metadata.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	out, err := exec.Command("gt", "log", "short", "--no-interactive").Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	var entries []GraphiteEntry

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		clean := ansiRegex.ReplaceAllString(line, "")
		name, prefix, isCurrent, circleBytePos := parseLine(clean)
		if name == "" {
			continue
		}
		// Column index from circle byte position (each column is 4 bytes apart: 3-byte char + space)
		colIndex := circleBytePos / 4

		entries = append(entries, GraphiteEntry{
			Name:       name,
			Prefix:     prefix,
			IsCurrent:  isCurrent,
			StackIndex: colIndex,
		})
	}

	if len(entries) == 0 {
		return nil
	}

	// Mark trunk (last entry in gt ls)
	entries[len(entries)-1].IsTrunk = true

	return entries
}

// parseLine extracts branch name, tree prefix, current status, and circle byte position.
func parseLine(line string) (name, prefix string, isCurrent bool, circleBytePos int) {
	treeChars := "│├└─┤┬┐┴○●◉◯┘┌"
	nameStart := -1
	circleBytePos = -1
	bytePos := 0
	for bytePos < len(line) {
		r, size := utf8.DecodeRuneInString(line[bytePos:])
		if r == '◉' || r == '●' {
			isCurrent = true
			circleBytePos = bytePos
		} else if r == '◯' || r == '○' {
			if circleBytePos < 0 {
				circleBytePos = bytePos
			}
		}
		if r != ' ' && !strings.ContainsRune(treeChars, r) {
			nameStart = bytePos
			break
		}
		bytePos += size
	}
	if nameStart < 0 {
		return "", "", false, 0
	}

	rest := line[nameStart:]
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return "", "", false, 0
	}

	return fields[0], line[:nameStart], isCurrent, circleBytePos
}

// ApplyGraphiteOrder reorders branches to match gt ls output and sets
// Graphite metadata on each branch.
func ApplyGraphiteOrder(branches []*models.Branch, entries []GraphiteEntry) {
	if len(entries) == 0 {
		return
	}

	branchByName := make(map[string]*models.Branch)
	for _, b := range branches {
		branchByName[b.Name] = b
	}

	for _, e := range entries {
		if b, ok := branchByName[e.Name]; ok {
			b.GraphiteTracked = true
			b.GraphitePrefix = e.Prefix
			b.GraphiteStackIndex = e.StackIndex
			// Clear the "  *" recency for HEAD - the filled circle already indicates it
			if b.Head {
				b.Recency = ""
			}
		}
	}

	// Build final order: gt ls order, then untracked at the end.
	placed := make(map[string]bool)
	result := make([]*models.Branch, 0, len(branches))

	for _, e := range entries {
		if b, ok := branchByName[e.Name]; ok && !placed[b.Name] {
			result = append(result, b)
			placed[b.Name] = true
		}
	}

	for _, b := range branches {
		if !placed[b.Name] {
			result = append(result, b)
			placed[b.Name] = true
		}
	}

	copy(branches, result)
}
