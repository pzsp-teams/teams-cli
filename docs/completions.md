# Shell completions

Teams CLI uses Cobra, so you can generate shell completions with the built-in
command.

## Bash completion

1. Generate completion script

   ```bash
   ./teams-cli completion bash > teams-cli-completion.bash
   ```

1. Install system-wide (recommended)

   ```bash
   sudo cp teams-cli-completion.bash /etc/bash_completion.d/
   ```

1. Or install user-only

   ```bash
   mkdir -p ~/.local/share/bash-completion/completions
   cp teams-cli-completion.bash ~/.local/share/bash-completion/completions/teams-cli
   ```

1. Reload your shell

   ```bash
   source ~/.bashrc
   ```

## Zsh completion

1. Generate completion script

   ```bash
   ./teams-cli completion zsh > _teams-cli
   ```

1. Install in Zsh function path

   ```bash
   sudo mkdir -p /usr/local/share/zsh/site-functions
   sudo cp _teams-cli /usr/local/share/zsh/site-functions/
   ```

1. Or install user-only

   ```bash
   mkdir -p ~/.zsh/completions
   cp _teams-cli ~/.zsh/completions/
   echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
   echo 'autoload -U compinit && compinit' >> ~/.zshrc
   ```

1. Reload your shell

   ```bash
   source ~/.zshrc
   ```

## Fish completion

1. Generate and install completion

   ```bash
   ./teams-cli completion fish > ~/.config/fish/completions/teams-cli.fish
   ```

1. Reload Fish configuration

   ```bash
   source ~/.config/fish/config.fish
   ```

## PowerShell completion

1. Generate completion script

   ```powershell
   ./teams-cli completion powershell > teams-cli.ps1
   ```

1. Create PowerShell profile directory

   ```powershell
   mkdir -p (Split-Path $PROFILE)
   ```

1. Add to PowerShell profile

   ```powershell
   Add-Content $PROFILE ". /path/to/teams-cli.ps1"
   ```

1. Reload PowerShell

   ```powershell
   & $PROFILE
   ```
