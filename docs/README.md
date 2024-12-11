# 🦉 gh-prowl

## Installation

To install this GitHub CLI extension, use the following command.

```shell
gh extension install kmtym1998/gh-prowl
```

## Features

- Monitors the status of checks for a Pull Request or ref (branch, commit hash, tag).
- Receive audio notifications when the checks are complete.

## Usage

Run the following command to start monitoring checks:

```bash
gh prowl [flags]
```

### Flags

- `-c, --current-branch`: Monitor the latest check status of the pull request linked to the current branch.
- `-b, --branch <branch>`: Monitor the latest check status of the specified branch.

If no PR is linked to the current branch, the extension will prompt you to select one from a list.

## Example

```bash
gh prowl --current-branch
```

To monitor checks for a specific branch:

```bash
gh prowl --branch feature-branch
```

Once all checks are complete, the results will be displayed in the terminal, and you'll receive a notification.
