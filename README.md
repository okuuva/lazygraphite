# lazygraphite

A terminal UI for [Graphite](https://graphite.dev) stack-based development, built on top of [lazygit](https://github.com/jesseduffield/lazygit).

## Acknowledgments

**This project would not exist without the tremendous work of [Jesse Duffield](https://github.com/jesseduffield) and all the [contributors](https://github.com/jesseduffield/lazygit/graphs/contributors) and [sponsors](https://github.com/sponsors/jesseduffield) of lazygit.**
Lazygit is the the best way of using git, and lazygraphite is a small fork that adds Graphite integration on top of it.

For everything lazygit can do -- staging individual lines, interactive rebasing, cherry-picking, bisecting, worktrees, custom commands, undo/redo, and much more -- please see the [lazygit README](https://github.com/jesseduffield/lazygit/blob/master/README.md).

## What lazygraphite adds

Lazygraphite integrates [Graphite's](https://graphite.dev) `gt` CLI into the lazygit TUI, giving you stack-aware workflows without leaving the terminal UI.

### Stack visualization

The branches panel displays your Graphite stacks as a metro-map tree, with colored columns per stack and the active stack placed far left for visibility.

### Graphite commands & keybindings

All commands are available from the relevant panels (files, branches, commits) via keybindings.

Two presets are available: **`"default"`** (lazygit-style) and **`"gt"`** (matches `gt` CLI commands).

> [!NOTE]
> The keybindings are still subject to change.
> While these might make sense in the context of their commit,
> adding support for new graphite commands might require some shuffling.
> If you like these, you can always save them as custom keybindings to keep them stable.
> And if you don't like them, please file an issue or a PR :)


| Action         | Default preset | GT preset | Description                                  |
|----------------|----------------|-----------|----------------------------------------------|
| **modify**     | `A`            | `m`       | Amend staged changes to a Graphite branch    |
| **create**     | `C`            | `c`       | Create a new Graphite branch                 |
| **get**        | `f`            | `g`       | Fetch/checkout a specific Graphite branch    |
| **reword**     | `r`            | `e`       | Edit a commit message via `gt modify --edit` |
| **checkout**   | `<space>`      | `<space>` | Checkout a branch in the stack               |
| **pr**         | `G`            | `p`       | Open the PR page for a branch                |
| **submit**     | `o`            | `s`       | Submit branch(es) for review                 |
| **submit-alt** | `P`            | —         | Alternative submit binding                   |
| **top**        | `<home>`       | `t`       | Jump to top of stack (`gt top`)              |
| **up**         | `<c-k>`        | `u`       | Move up in stack (`gt up`)                   |
| **down**       | `<c-j>`        | `d`       | Move down in stack (`gt down`)               |
| **restack**    | `r`            | `r`       | Restack (`gt restack`)                       |
| **sync**       | `p`            | `S`       | Sync with remote (`gt sync --force`)         |
| **undo**       | `z`            | `U`       | Undo last Graphite operation (`gt undo`)     |

When moving up or jumping to top, if multiple children/leaf branches exist, you'll be prompted to choose.

## Prerequisites

- [lazygit prerequisites](https://github.com/jesseduffield/lazygit/blob/master/README.md#installation) (Go, git, etc.)
- The [Graphite CLI](https://graphite.dev/docs/installing-the-cli) (`gt`) must be installed and in your PATH
- A Graphite-initialized repository (`.graphite_metadata.db` must exist)

## Installation

Build from source:

```bash
git clone https://github.com/YOUR_ORG/lazygraphite.git
cd lazygraphite
make build
```

## Configuration

Enable Graphite integration in your lazygit config file (`~/.config/lazygit/config.yml`):

```yaml
graphite:
  enabled: true
  keymap: "default"       # "default" or "gt"
  commandLogSize: 20      # Size of command log panel (0 = use gui.commandLogSize)
  commandLogTimeout: 3    # Seconds to keep command log enlarged after gt commands
```

Individual keys can be overridden regardless of preset via `keybinding.graphite` in your config file.
See the [Keybindings](#keybindings) table above for all available keys and their defaults.

## Contributing

Contributions are welcome! For general lazygit development, see the [lazygit contributing guide](https://github.com/jesseduffield/lazygit/blob/master/CONTRIBUTING.md).
For Graphite-specific features, open an issue or PR in this repository.
