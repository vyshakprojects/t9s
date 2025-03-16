# MyTunnel - Interactive SSH Tunneling Manager

MyTunnel is a CLI tool that provides a k9s-like interface for managing SSH tunnels through bastion servers.

## Features

- Interactive terminal UI with Vim-style navigation
- Bastion server configuration management
- Real-time port monitoring and tunnel management
- SSH key and password authentication support

## Installation

```bash
go install github.com/yourusername/mytunnel@latest
```

## Configuration

Create a configuration file at `~/.mytunnel/config.yaml`:

```yaml
bastions:
  my-bastion:
    host: bastion.example.com
    user: username
    port: 22
    auth_type: key  # or password
    key_path: ~/.ssh/id_rsa
```

## Usage

Basic commands:

- `mytunnel` - Opens the interactive UI
- `mytunnel list-bastions` - Shows available bastions
- `mytunnel add-bastion --name my-bastion ...` - Adds a bastion server
- `mytunnel --bastion my-bastion` - Launches UI for specific bastion

## Navigation

In the interactive UI:

- `j/k` - Navigate through port list
- `Enter/Space` - Start SSH tunneling for selected port
- `t` - Toggle to view active tunnels
- `d` - Delete/close a tunnel
- `/` - Search/filter available ports
- `:q/esc` - Quit

## License

MIT 