package helpers

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

type GraphiteHelper struct {
	c *HelperCommon
}

func NewGraphiteHelper(c *HelperCommon) *GraphiteHelper {
	return &GraphiteHelper{
		c: c,
	}
}

// RunAndStream runs a gt command with output streamed to the Command Log panel
// instead of suspending the TUI for a subprocess terminal.
func (self *GraphiteHelper) RunAndStream(cmdObj *oscommands.CmdObj, waitingStatus string) error {
	return self.c.WithWaitingStatus(waitingStatus, func(gocui.Task) error {
		if err := cmdObj.StreamOutput().Run(); err != nil {
			return fmt.Errorf(
				self.c.Tr.GitCommandFailed, self.c.UserConfig().Keybinding.Universal.ExtrasMenu,
			)
		}
		self.c.Refresh(types.RefreshOptions{Mode: types.ASYNC})
		return nil
	})
}

// AmendToCommit amends staged changes to the commit pointed to by a Graphite branch.
// It checks out the target branch, applies the staged changes, runs gt modify,
// then returns to the original branch and restores unstaged changes.
func (self *GraphiteHelper) AmendToCommit(commit *models.Commit, currentBranchName string) error {
	branchAtCommit, err := self.FindBranchAtCommit(commit.Hash())
	if err != nil {
		return err
	}
	if branchAtCommit == "" {
		return fmt.Errorf("%s", self.c.Tr.GraphiteNoBranchAtCommit)
	}

	script := fmt.Sprintf(`bash -c '
set -e

TARGET_BRANCH="%s"
ORIGINAL_BRANCH="%s"

# Check if we have staged changes
if ! git diff --cached --quiet; then
  HAS_STAGED=1
else
  HAS_STAGED=0
fi

# If we have staged changes, save them as a patch first
PATCH_FILE=""
if [ "$HAS_STAGED" -eq 1 ]; then
  PATCH_FILE=$(mktemp)
  git diff --cached > "$PATCH_FILE"
fi

# Stash ALL changes (staged + unstaged + untracked)
STASH_NEEDED=0
if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
  STASH_NEEDED=1
  echo "Stashing all local changes..."
  git stash push --include-untracked -m "lazygit: gt modify autostash"
fi

# Checkout target branch
echo "Checking out branch: $TARGET_BRANCH"
gt checkout "$TARGET_BRANCH"

# Apply the patch if we had staged changes
if [ "$HAS_STAGED" -eq 1 ]; then
  echo "Applying staged changes..."
  if ! git apply --3way --index "$PATCH_FILE" 2>&1; then
    if git diff --name-only --diff-filter=U | grep -q .; then
      echo ""
      echo "Patch applied with conflicts. Please resolve them."
    else
      echo ""
      echo "ERROR: Patch could not be applied."
      rm -f "$PATCH_FILE"
      git reset --hard HEAD
      gt checkout "$ORIGINAL_BRANCH"
      if [ "$STASH_NEEDED" -eq 1 ]; then
        git stash pop
      fi
      echo "Press enter to continue..."
      read
      exit 1
    fi
  fi
  rm -f "$PATCH_FILE"
fi

# Run gt modify
echo "Running gt modify..."
gt modify

# Return to original branch
echo "Returning to branch: $ORIGINAL_BRANCH"
gt checkout "$ORIGINAL_BRANCH"

# Pop stash if we stashed
if [ "$STASH_NEEDED" -eq 1 ]; then
  echo "Restoring unstaged changes..."
  git stash pop
fi

echo "Done! Staged changes have been amended to commit on $TARGET_BRANCH"
'`, branchAtCommit, currentBranchName)

	self.c.LogAction(self.c.Tr.Actions.GraphiteModify)
	cmdObj := self.c.OS().Cmd.NewShell(script, "")
	return self.c.RunSubprocessAndRefresh(cmdObj)
}

// RewordCommit checks out the branch at the given commit, runs gt modify --edit,
// then returns to the top of the stack.
func (self *GraphiteHelper) RewordCommit(commit *models.Commit) error {
	branchAtCommit, err := self.FindBranchAtCommit(commit.Hash())
	if err != nil {
		return err
	}
	if branchAtCommit == "" {
		return fmt.Errorf("%s", self.c.Tr.GraphiteNoBranchAtCommit)
	}

	return self.RewordBranch(branchAtCommit)
}

// RewordBranch checks out the given branch, runs gt modify --edit,
// then returns to the top of the stack.
func (self *GraphiteHelper) RewordBranch(branchName string) error {
	script := fmt.Sprintf(`bash -c '
set -e

TARGET_BRANCH="%s"

echo "Checking out branch: $TARGET_BRANCH"
gt checkout "$TARGET_BRANCH"

echo "Running gt modify --edit..."
gt modify --edit --no-interactive

echo "Returning to top of stack..."
gt top

echo "Done!"
'`, branchName)

	self.c.LogAction(self.c.Tr.Actions.GraphiteReword)
	cmdObj := self.c.OS().Cmd.NewShell(script, "")
	return self.c.RunSubprocessAndRefresh(cmdObj)
}

// CheckoutCommitBranch checks out the Graphite branch that points at the given commit.
func (self *GraphiteHelper) CheckoutCommitBranch(commit *models.Commit) error {
	branchAtCommit, err := self.FindBranchAtCommit(commit.Hash())
	if err != nil {
		return err
	}
	if branchAtCommit == "" {
		return fmt.Errorf("%s", self.c.Tr.GraphiteNoBranchAtCommit)
	}

	self.c.LogAction(self.c.Tr.Actions.GraphiteCheckout)
	return self.RunAndStream(self.c.Git().Graphite.CheckoutCmdObj(branchAtCommit), "Checking out...")
}

// PrForCommit opens the PR page for the branch at the given commit.
func (self *GraphiteHelper) PrForCommit(commit *models.Commit) error {
	branchAtCommit, err := self.FindBranchAtCommit(commit.Hash())
	if err != nil {
		return err
	}
	if branchAtCommit == "" {
		return fmt.Errorf("%s", self.c.Tr.GraphiteNoBranchAtCommit)
	}

	self.c.LogAction(self.c.Tr.Actions.GraphitePr)
	return self.RunAndStream(self.c.Git().Graphite.PrCmdObj(branchAtCommit), "Opening PR...")
}

// FindBranchAtCommit returns the name of the branch that points directly at the given commit hash.
func (self *GraphiteHelper) FindBranchAtCommit(hash string) (string, error) {
	cmdStr := fmt.Sprintf("git branch --points-at %s --format=%%(refname:short)", hash)
	output, err := self.c.Git().Custom.RunWithOutput(cmdStr)
	if err != nil {
		return "", err
	}

	for _, name := range utils.SplitLines(output) {
		name = strings.TrimSpace(name)
		if name != "" && !strings.HasPrefix(name, "(HEAD") {
			return name, nil
		}
	}

	return "", nil
}
