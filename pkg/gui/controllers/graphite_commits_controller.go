package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type GraphiteCommitsController struct {
	baseController
	*ListControllerTrait[*models.Commit]
	c *ControllerCommon
}

var _ types.IController = &GraphiteCommitsController{}

func NewGraphiteCommitsController(c *ControllerCommon) *GraphiteCommitsController {
	return &GraphiteCommitsController{
		baseController: baseController{},
		ListControllerTrait: NewListControllerTrait(
			c,
			c.Contexts().LocalCommits,
			c.Contexts().LocalCommits.GetSelected,
			c.Contexts().LocalCommits.GetSelectedItems,
		),
		c: c,
	}
}

func (self *GraphiteCommitsController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	graphiteDisabled := self.graphiteDisabledReason()

	return []*types.Binding{
		{
			Key:               opts.GetKey(opts.Config.Graphite.Modify),
			Handler:           self.withItem(self.amendToCommit),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteModify,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Reword),
			Handler:           self.withItem(self.reword),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteReword,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Checkout),
			Handler:           self.withItem(self.checkout),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteCheckout,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Pr),
			Handler:           self.withItem(self.pr),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphitePr,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Submit),
			Handler:           self.submit,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteSubmit,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.SubmitAlt),
			Handler:           self.submit,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteSubmit,
			Alternative:       opts.Config.Graphite.Submit,
		},
	}
}

func (self *GraphiteCommitsController) amendToCommit(commit *models.Commit) error {
	currentBranch, err := self.c.Git().Branch.CurrentBranchName()
	if err != nil {
		return err
	}
	return self.c.Helpers().Graphite.AmendToCommit(commit, currentBranch)
}

func (self *GraphiteCommitsController) reword(commit *models.Commit) error {
	return self.c.Helpers().Graphite.RewordCommit(commit)
}

func (self *GraphiteCommitsController) checkout(commit *models.Commit) error {
	return self.c.Helpers().Graphite.CheckoutCommitBranch(commit)
}

func (self *GraphiteCommitsController) pr(commit *models.Commit) error {
	return self.c.Helpers().Graphite.PrForCommit(commit)
}

func (self *GraphiteCommitsController) submit() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteSubmit)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.SubmitCmdObj(), "Submitting...")
}

func (self *GraphiteCommitsController) graphiteDisabledReason() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		if !self.c.Git().Graphite.Enabled() {
			return &types.DisabledReason{Text: self.c.Tr.GraphiteNotEnabled}
		}
		return nil
	}
}

func (self *GraphiteCommitsController) Context() types.Context {
	return self.c.Contexts().LocalCommits
}
