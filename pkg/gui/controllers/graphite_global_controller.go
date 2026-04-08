package controllers

import (
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
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.RestackCmdObj())
}

func (self *GraphiteGlobalController) sync() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteSync)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.SyncCmdObj())
}

func (self *GraphiteGlobalController) undo() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteUndo)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.UndoCmdObj())
}

func (self *GraphiteGlobalController) top() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteTop)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.TopCmdObj())
}

func (self *GraphiteGlobalController) up() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteUp)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.UpCmdObj())
}

func (self *GraphiteGlobalController) down() error {
	self.c.LogAction(self.c.Tr.Actions.GraphiteDown)
	return self.c.RunSubprocessAndRefresh(self.c.Git().Graphite.DownCmdObj())
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
