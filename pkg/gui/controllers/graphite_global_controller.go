package controllers

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type GraphiteGlobalController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &GraphiteGlobalController{}

func NewGraphiteGlobalController(c *ControllerCommon) *GraphiteGlobalController {
	return &GraphiteGlobalController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *GraphiteGlobalController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	graphiteDisabled := self.graphiteDisabledReason()

	return []*types.Binding{
		{
			Key:               opts.GetKey(opts.Config.Graphite.Restack),
			Handler:           self.restack,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteRestack,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Sync),
			Handler:           self.sync,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteSync,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Undo),
			Handler:           self.undo,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteUndo,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Top),
			Handler:           self.top,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteTop,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Up),
			Handler:           self.up,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteUp,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Down),
			Handler:           self.down,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteDown,
		},
	}
}

func (self *GraphiteGlobalController) restack() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteRestack)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.RestackCmdObj(), "Restacking...")
}

func (self *GraphiteGlobalController) sync() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteSync)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.SyncCmdObj(), "Syncing...")
}

func (self *GraphiteGlobalController) undo() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteUndo)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.UndoCmdObj(), "Undoing...")
}

func (self *GraphiteGlobalController) top() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteTop)

	// Find all leaf (top) branches reachable from the current branch.
	// If there are multiple, let the user pick which top to go to.
	currentBranch := self.c.Model().CheckedOutBranch
	leaves := self.c.Git().Graphite.LeafBranches(currentBranch)
	if len(leaves) <= 1 {
		return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.TopCmdObj(), "Going to top...")
	}

	menuItems := make([]*types.MenuItem, len(leaves))
	for i, leaf := range leaves {
		menuItems[i] = &types.MenuItem{
			Label: leaf,
			OnPress: func() error {
				return self.c.Helpers().Graphite.RunAndStream(
					self.c.Git().Graphite.CheckoutCmdObj(leaf), "Going to top...")
			},
		}
	}

	return self.c.Menu(types.CreateMenuOptions{
		Title: self.c.Tr.GraphiteSelectTop,
		Items: menuItems,
	})
}

func (self *GraphiteGlobalController) up() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteUp)

	// Check if there are multiple children; if so, prompt the user to pick one
	output, err := self.c.Git().Graphite.ChildrenCmdObj().RunWithOutput()
	if err != nil {
		// If gt children fails, fall back to regular gt up
		return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.UpCmdObj(), "Going up...")
	}

	children := filterEmpty(strings.Split(output, "\n"))
	if len(children) <= 1 {
		return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.UpCmdObj(), "Going up...")
	}

	menuItems := make([]*types.MenuItem, len(children))
	for i, child := range children {
		menuItems[i] = &types.MenuItem{
			Label: child,
			OnPress: func() error {
				return self.c.Helpers().Graphite.RunAndStream(
					self.c.Git().Graphite.UpToCmdObj(child), "Going up...")
			},
		}
	}

	return self.c.Menu(types.CreateMenuOptions{
		Title: self.c.Tr.GraphiteSelectChild,
		Items: menuItems,
	})
}

func filterEmpty(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (self *GraphiteGlobalController) down() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteDown)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.DownCmdObj(), "Going down...")
}

func (self *GraphiteGlobalController) graphiteDisabledReason() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		if !self.c.Git().Graphite.Enabled() {
			return &types.DisabledReason{Text: self.c.Tr.GraphiteNotEnabled}
		}
		return nil
	}
}

func (self *GraphiteGlobalController) Context() types.Context {
	return self.c.Contexts().Global
}
