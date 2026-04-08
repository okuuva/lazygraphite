package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type GraphiteFilesController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &GraphiteFilesController{}

func NewGraphiteFilesController(c *ControllerCommon) *GraphiteFilesController {
	return &GraphiteFilesController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *GraphiteFilesController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	graphiteDisabled := self.graphiteDisabledReason()

	return []*types.Binding{
		{
			Key:               opts.GetKey(opts.Config.Graphite.Modify),
			Handler:           self.modify,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteModify,
		},
		{
			Key:               opts.GetKey(opts.Config.Graphite.Create),
			Handler:           self.create,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphiteCreate,
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
		{
			Key:               opts.GetKey(opts.Config.Graphite.Pr),
			Handler:           self.pr,
			GetDisabledReason: graphiteDisabled,
			Description:       self.c.Tr.GraphitePr,
		},
	}
}

func (self *GraphiteFilesController) modify() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteModify)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.ModifyCmdObj(), "Modifying...")
}

func (self *GraphiteFilesController) create() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteCreate)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.CreateCmdObj())
}

func (self *GraphiteFilesController) submit() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteSubmit)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.SubmitCmdObj(), "Submitting...")
}

func (self *GraphiteFilesController) pr() error {
	self.c.LogAction(self.c.Tr.Actions.GraphitePr)
	return self.c.Helpers().Graphite.RunAndStream(self.c.Git().Graphite.PrCmdObj(""), "Opening PR...")
}

func (self *GraphiteFilesController) graphiteDisabledReason() func() *types.DisabledReason {
	return func() *types.DisabledReason {
		if !self.c.Git().Graphite.Enabled() {
			return &types.DisabledReason{Text: self.c.Tr.GraphiteNotEnabled}
		}
		return nil
	}
}

func (self *GraphiteFilesController) Context() types.Context {
	return self.c.Contexts().Files
}
