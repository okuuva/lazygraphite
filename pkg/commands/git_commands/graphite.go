package git_commands

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
)

type GraphiteCommands struct {
	*GitCommon
}

func NewGraphiteCommands(gitCommon *GitCommon) *GraphiteCommands {
	return &GraphiteCommands{
		GitCommon: gitCommon,
	}
}

func (self *GraphiteCommands) Enabled() bool {
	if !self.UserConfig().Graphite.Enabled {
		return false
	}
	_, err := exec.LookPath("gt")
	return err == nil
}

func (self *GraphiteCommands) ModifyCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "modify"})
}

func (self *GraphiteCommands) ModifyEditCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "modify", "--edit", "--no-interactive"})
}

func (self *GraphiteCommands) CreateCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "create"})
}

func (self *GraphiteCommands) GetCmdObj(branch string) *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "get", branch})
}

func (self *GraphiteCommands) PrCmdObj(branch string) *oscommands.CmdObj {
	if branch == "" {
		return self.os.Cmd.New([]string{"gt", "pr"})
	}
	return self.os.Cmd.New([]string{"gt", "pr", branch})
}

func (self *GraphiteCommands) SubmitCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "submit", "--restack", "--no-edit", "--draft"})
}

func (self *GraphiteCommands) RestackCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "restack"})
}

func (self *GraphiteCommands) SyncCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "sync", "--force"})
}

func (self *GraphiteCommands) UndoCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "undo"})
}

func (self *GraphiteCommands) CheckoutCmdObj(branch string) *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "checkout", branch})
}

func (self *GraphiteCommands) TopCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "top"})
}

func (self *GraphiteCommands) UpCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "up"})
}

func (self *GraphiteCommands) UpToCmdObj(branch string) *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "up", "--to", branch})
}

func (self *GraphiteCommands) ChildrenCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "children"})
}

func (self *GraphiteCommands) DownCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "down"})
}

// LeafBranches returns all leaf (top) branches reachable from the given branch
// by reading the graphite metadata DB. Returns nil on any error.
func (self *GraphiteCommands) LeafBranches(from string) []string {
	dbPath := filepath.Join(self.repoPaths.WorktreeGitDirPath(), ".graphite_metadata.db")
	out, err := exec.Command("sqlite3", "-json", dbPath,
		"SELECT branch_name, parent_branch_name FROM branch_metadata").Output()
	if err != nil {
		return nil
	}

	var rows []struct {
		BranchName       string  `json:"branch_name"`
		ParentBranchName *string `json:"parent_branch_name"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil
	}

	childrenOf := make(map[string][]string)
	for _, row := range rows {
		if row.ParentBranchName != nil && *row.ParentBranchName != "" {
			childrenOf[*row.ParentBranchName] = append(childrenOf[*row.ParentBranchName], row.BranchName)
		}
	}

	// DFS to find all leaves reachable from `from`
	var leaves []string
	var walk func(name string)
	walk = func(name string) {
		kids := childrenOf[name]
		if len(kids) == 0 {
			leaves = append(leaves, name)
			return
		}
		for _, kid := range kids {
			walk(kid)
		}
	}

	for _, kid := range childrenOf[from] {
		walk(kid)
	}

	slices.Sort(leaves)
	return leaves
}

