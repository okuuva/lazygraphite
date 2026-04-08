package git_commands

import (
	"os/exec"

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

func (self *GraphiteCommands) DownCmdObj() *oscommands.CmdObj {
	return self.os.Cmd.New([]string{"gt", "down"})
}
