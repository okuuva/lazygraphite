package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type GraphiteBranchesController struct {
	baseController
	*ListControllerTrait[*models.Branch]
	c *ControllerCommon
}

var _ types.IController = &GraphiteBranchesController{}

func NewGraphiteBranchesController(c *ControllerCommon) *GraphiteBranchesController {
	return &GraphiteBranchesController{
		baseController: baseController{},
		ListControllerTrait: NewListControllerTrait(
			c,
			c.Contexts().Branches,
			c.Contexts().Branches.GetSelected,
			c.Contexts().Branches.GetSelectedItems,
		),
		c: c,
	}
}

func (self *GraphiteBranchesController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	graphiteDisabled := self.graphiteDisabledReason()

	return []*types.Binding{
		{
			Key:               opts.GetKey(opts.Config.Graphite.Reword),
			Handler:           self.withItem(self.reword),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteReword,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Get),
			Handler:           self.withItem(self.get),
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteGet,
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

func (self *GraphiteBranchesController) reword(branch *models.Branch) error {
	return self.c.Helpers().Graphite.RewordBranch(branch.Name)
}

func (self *GraphiteBranchesController) get(branch *models.Branch) error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteGet)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.GetCmdObj(branch.Name), "Getting branch...")
}

func (self *GraphiteBranchesController) checkout(branch *models.Branch) error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteCheckout)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.CheckoutCmdObj(branch.Name), "Checking out...")
}

func (self *GraphiteBranchesController) pr(branch *models.Branch) error {
	self.c.LogAction(self.c.Tr.Actions.GraphitePr)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.PrCmdObj(branch.Name), "Opening PR...")
}

func (self *GraphiteBranchesController) submit() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteSubmit)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.SubmitCmdObj(), "Submitting...")
}

func (self *GraphiteBranchesController) graphiteDisabledReason() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		if !self.c.Git().Graphite.Enabled() {
			return &types.DisabledReason{Text: self.c.Tr.GraphiteNotEnabled}
		}
		return nil
	}
}

func (self *GraphiteBranchesController) Context() types.Context {
	return self.c.Contexts().Branches
}
