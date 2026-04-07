package git_commands

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

type GraphiteEntry struct {
	Name       string
	Prefix     string
	StackIndex int // column index for coloring
}

type graphiteRow struct {
	BranchName       string  `json:"branch_name"`
	ParentBranchName *string `json:"parent_branch_name"`
	ValidationResult *string `json:"validation_result"`
}

// metroEntry is an intermediate representation for metro map generation.
type metroEntry struct {
	name        string
	column      int
	isCurrent   bool
	isTrunk     bool
	connectCols []int // columns of non-primary children (for branch point connectors)
}

// colAllocator manages column assignment with reuse.
type colAllocator struct {
	next  int
	freed []int
}

func (a *colAllocator) alloc() int {
	if len(a.freed) > 0 {
		col := a.freed[0]
		a.freed = a.freed[1:]
		return col
	}
	c := a.next
	a.next++
	return c
}

func (a *colAllocator) free(col int) {
	a.freed = append(a.freed, col)
	slices.Sort(a.freed)
}

// LoadGraphiteEntries builds the metro map from the SQLite metadata DB.
func LoadGraphiteEntries(repoGitDir string) []GraphiteEntry {
	dbPath := filepath.Join(repoGitDir, ".graphite_metadata.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	out, err := exec.Command("sqlite3", "-json", dbPath,
		"SELECT branch_name, parent_branch_name, validation_result FROM branch_metadata").Output()
	if err != nil {
		return nil
	}

	var rows []graphiteRow
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil
	}

	// Get local branch names to filter out stale DB entries
	localOut, err := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/").Output()
	if err != nil {
		return nil
	}
	localBranches := make(map[string]bool)
	for _, name := range strings.Split(strings.TrimSpace(string(localOut)), "\n") {
		if name != "" {
			localBranches[name] = true
		}
	}

	parentOf := make(map[string]string)
	childrenOf := make(map[string][]string)
	trunkName := ""
	tracked := make(map[string]bool)

	for _, row := range rows {
		if !localBranches[row.BranchName] {
			continue // skip DB entries for deleted branches
		}
		tracked[row.BranchName] = true
		if row.ParentBranchName != nil && *row.ParentBranchName != "" {
			parentOf[row.BranchName] = *row.ParentBranchName
		}
		if row.ValidationResult != nil && *row.ValidationResult == "TRUNK" {
			trunkName = row.BranchName
		}
	}

	if trunkName == "" {
		return nil
	}

	// Build children map from parent pointers (only for tracked local branches)
	for name, parent := range parentOf {
		if tracked[parent] {
			childrenOf[parent] = append(childrenOf[parent], name)
		}
	}
	for k := range childrenOf {
		slices.Sort(childrenOf[k])
	}

	// Find HEAD branch
	headOut, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	headBranch := ""
	if err == nil {
		headBranch = strings.TrimSpace(string(headOut))
	}

	// DFS from trunk to build ordered metro entries
	alloc := &colAllocator{next: 1}
	entries := metroDFS(trunkName, 0, alloc, childrenOf, headBranch)

	// Mark trunk
	if len(entries) > 0 {
		entries[len(entries)-1].isTrunk = true
	}

	// Generate prefixes
	return generatePrefixes(entries)
}

// metroDFS performs a DFS to assign columns and ordering matching gt ls layout.
// Primary child (first alphabetically) stays at parent's column.
// Other children get new columns, allocated after primary subtree is processed
// (allowing column reuse).
func metroDFS(name string, col int, alloc *colAllocator,
	childrenOf map[string][]string, headBranch string,
) []metroEntry {
	kids := childrenOf[name]
	if len(kids) == 0 {
		return []metroEntry{{name: name, column: col, isCurrent: name == headBranch}}
	}

	primary := kids[0]
	others := kids[1:]

	// Process primary subtree first (appears at top of display)
	result := metroDFS(primary, col, alloc, childrenOf, headBranch)

	// Allocate columns for non-primary children (after primary frees any it used)
	connectCols := make([]int, len(others))
	for i := range others {
		connectCols[i] = alloc.alloc()
	}

	// Process non-primary subtrees
	for i, other := range others {
		result = append(result, metroDFS(other, connectCols[i], alloc, childrenOf, headBranch)...)
	}

	// Free columns (they're closed at this branch point)
	for _, c := range connectCols {
		alloc.free(c)
	}

	// Add this entry
	result = append(result, metroEntry{
		name:        name,
		column:      col,
		isCurrent:   name == headBranch,
		connectCols: connectCols,
	})

	return result
}

// generatePrefixes computes the tree-drawing prefix for each metro entry.
func generatePrefixes(entries []metroEntry) []GraphiteEntry {
	if len(entries) == 0 {
		return nil
	}

	// Find max column used
	maxCol := 0
	for _, e := range entries {
		if e.column > maxCol {
			maxCol = e.column
		}
		for _, c := range e.connectCols {
			if c > maxCol {
				maxCol = c
			}
		}
	}

	// Compute active ranges for each column.
	// A column C is "active" (shows │) between its first circle and its closing connector.
	// With column reuse, there are multiple ranges per column.
	type lineRange struct{ top, bottom int }
	colRanges := make(map[int][]lineRange)

	// Collect circle lines and connector lines per column
	circleLines := make(map[int][]int) // col → line indices of circles
	connectorLines := make(map[int][]int) // col → line indices of branch points closing this col

	for i, e := range entries {
		circleLines[e.column] = append(circleLines[e.column], i)
		for _, c := range e.connectCols {
			connectorLines[c] = append(connectorLines[c], i)
		}
	}

	// Match circles to connectors for each column
	for col, connLines := range connectorLines {
		circs := circleLines[col]
		connIdx := 0
		circIdx := 0
		for connIdx < len(connLines) {
			connLine := connLines[connIdx]
			firstCirc := -1
			for circIdx < len(circs) && circs[circIdx] < connLine {
				if firstCirc < 0 {
					firstCirc = circs[circIdx]
				}
				circIdx++
			}
			if firstCirc >= 0 {
				colRanges[col] = append(colRanges[col], lineRange{firstCirc, connLine})
			}
			connIdx++
		}
	}

	// Column 0 (primary chain) has no connector - its range spans all its circles
	if circs, ok := circleLines[0]; ok && len(circs) > 0 {
		colRanges[0] = []lineRange{{circs[0], circs[len(circs)-1]}}
	}

	// Helper: is column C active at line i?
	isActive := func(col, lineIdx int) bool {
		for _, r := range colRanges[col] {
			if lineIdx >= r.top && lineIdx <= r.bottom {
				return true
			}
		}
		return false
	}

	// Generate prefix string for each entry
	totalWidth := (maxCol+1)*2 + 1
	result := make([]GraphiteEntry, len(entries))

	for i, e := range entries {
		var prefix strings.Builder

		// Determine if this is a branch point with connectors
		hasConnect := len(e.connectCols) > 0
		lastConnect := -1
		connectSet := make(map[int]bool)
		if hasConnect {
			for _, c := range e.connectCols {
				connectSet[c] = true
				if c > lastConnect {
					lastConnect = c
				}
			}
		}

		for col := 0; col <= maxCol; col++ {
			isOwnCol := col == e.column
			inConnectRange := hasConnect && col > e.column && col <= lastConnect

			// Column character
			if isOwnCol {
				if e.isCurrent {
					prefix.WriteRune('◉')
				} else {
					prefix.WriteRune('◯')
				}
			} else if inConnectRange && connectSet[col] {
				if col == lastConnect {
					prefix.WriteRune('┘')
				} else {
					prefix.WriteRune('┴')
				}
			} else if inConnectRange {
				prefix.WriteRune('─')
			} else if isActive(col, i) && entries[i].column != col {
				prefix.WriteRune('│')
			} else {
				prefix.WriteRune(' ')
			}

			// Separator after column character
			if isOwnCol && hasConnect && col < lastConnect {
				prefix.WriteRune('─')
			} else if inConnectRange && col < lastConnect {
				prefix.WriteRune('─')
			} else {
				prefix.WriteRune(' ')
			}
		}

		// Pad to fixed width
		for prefix.Len() < totalWidth {
			prefix.WriteRune(' ')
		}

		result[i] = GraphiteEntry{
			Name:       e.name,
			Prefix:     prefix.String(),
			StackIndex: e.column,
		}
	}

	return result
}

// ApplyGraphiteOrder reorders branches to match the metro map and sets metadata.
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
			if b.Head {
				b.Recency = ""
			}
		}
	}

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
