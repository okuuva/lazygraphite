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
	Name          string
	Prefix        string
	StackIndex    int // color index for this entry's column
	StackPosition int // depth from trunk (0 = trunk, 1 = first branch, etc.)
}

type graphiteRow struct {
	BranchName       string  `json:"branch_name"`
	ParentBranchName *string `json:"parent_branch_name"`
	ValidationResult *string `json:"validation_result"`
}

// metroEntry is an intermediate representation for metro map generation.
type metroEntry struct {
	name          string
	column        int
	stackPosition int   // depth from trunk (0 = trunk, 1 = first branch, etc.)
	isCurrent     bool
	isTrunk       bool
	connectCols   []int // columns of non-primary children (for branch point connectors)
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

	// Build headPath: set of branches on the path from HEAD to trunk
	headPath := make(map[string]bool)
	for cur := headBranch; cur != "" && cur != trunkName; cur = parentOf[cur] {
		headPath[cur] = true
	}
	headPath[trunkName] = true

	// DFS from trunk to build ordered metro entries
	alloc := &colAllocator{next: 1}
	entries := metroDFS(trunkName, 0, 0, alloc, childrenOf, headBranch, headPath)

	// Mark trunk
	if len(entries) > 0 {
		entries[len(entries)-1].isTrunk = true
	}

	// Generate prefixes
	return generatePrefixes(entries)
}

// metroDFS performs a DFS to assign columns, ordering, and stack positions.
// The child on the HEAD path is made primary (stays at parent's column),
// falling back to the first alphabetically if no child is on the HEAD path.
// depth tracks the position within the column (0 = trunk/root of stack).
func metroDFS(name string, col int, depth int, alloc *colAllocator,
	childrenOf map[string][]string, headBranch string, headPath map[string]bool,
) []metroEntry {
	kids := childrenOf[name]
	if len(kids) == 0 {
		return []metroEntry{{name: name, column: col, stackPosition: depth, isCurrent: name == headBranch}}
	}

	// Pick primary child: prefer the one on HEAD's path, else first alphabetically
	primaryIdx := 0
	for i, kid := range kids {
		if headPath[kid] {
			primaryIdx = i
			break
		}
	}
	primary := kids[primaryIdx]
	others := make([]string, 0, len(kids)-1)
	for i, kid := range kids {
		if i != primaryIdx {
			others = append(others, kid)
		}
	}

	// Process primary subtree first (stays at same column, depth increments)
	result := metroDFS(primary, col, depth+1, alloc, childrenOf, headBranch, headPath)

	// Allocate and process non-primary children one at a time.
	// Non-primary children start new columns at depth 1.
	connectCols := make([]int, 0, len(others))
	for _, other := range others {
		c := alloc.alloc()
		connectCols = append(connectCols, c)
		result = append(result, metroDFS(other, c, 1, alloc, childrenOf, headBranch, headPath)...)
	}

	// Free columns (they're closed at this branch point)
	for _, c := range connectCols {
		alloc.free(c)
	}

	// Add this entry
	result = append(result, metroEntry{
		name:          name,
		column:        col,
		stackPosition: depth,
		isCurrent:     name == headBranch,
		connectCols:   connectCols,
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

		// Count visible markers before this entry's column to compute color index.
		// This must match the marker-counting logic in colorGraphitePrefix.
		colorIndex := 0
		for col := 0; col <= maxCol; col++ {
			isOwnCol := col == e.column
			inConnectRange := hasConnect && col > e.column && col <= lastConnect

			// Determine if this column produces a marker character
			isMarker := false
			if isOwnCol {
				isMarker = true
				if e.isCurrent {
					prefix.WriteRune('◉')
				} else {
					prefix.WriteRune('◯')
				}
			} else if inConnectRange && connectSet[col] {
				isMarker = true
				if col == lastConnect {
					prefix.WriteRune('┘')
				} else {
					prefix.WriteRune('┴')
				}
			} else if inConnectRange {
				prefix.WriteRune('─')
			} else if isActive(col, i) && entries[i].column != col {
				isMarker = true
				prefix.WriteRune('│')
			} else {
				prefix.WriteRune(' ')
			}

			if col < e.column && isMarker {
				colorIndex++
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
			Name:          e.name,
			Prefix:        prefix.String(),
			StackIndex:    colorIndex,
			StackPosition: e.stackPosition,
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
			b.GraphiteStackPosition = e.StackPosition
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
