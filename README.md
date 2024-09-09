# kvm-ssh: KVM SSH Connection and Port Forwarding Tool

`kvm-ssh` is a command-line tool designed to simplify SSH connections to KVM (Kernel-based Virtual Machine) instances and provide easy port forwarding capabilities.

## Features

- List all running KVM virtual machines
- Easily connect to a specific KVM virtual machine via SSH
- Forward ports from a KVM virtual machine to the local machine
- Sensible defaults for common options

## Notes

- This tool requires `sudo` access to run `virsh` commands.
- Ensure that the `ssh` command is available on your system.
- The tool reads VM information from `/var/lib/libvirt/dnsmasq/<bridge_name>.status`.

## Tested Environment

This tool has been confirmed to work in the following environment:

```
$ lsb_release -a
No LSB modules are available.
Distributor ID: Debian
Description:    Debian GNU/Linux trixie/sid
Release:        n/a
Codename:       trixie
```

## Installation

To install `kvm-ssh`, clone the repository and build the tool:

```bash
$ go build
```

Make sure to place the built `kvm-ssh` binary in your system's PATH for easy access.

## Usage

### General Syntax

The general syntax for using `kvm-ssh` is as follows:

```
kvm-ssh <command> [options] <arguments>
```

Where `<command>` is one of `list`, `connect`, or `forward`.

### Default Values

- The default bridge name is set to `virbr0`.
- The default user is set to the value of the `USER` environment variable.

You only need to specify `--bridge` or `--user` if you want to use a different value.

### List all running KVM virtual machines

```bash
$ kvm-ssh list
```

### Connect to a KVM virtual machine

```bash
$ kvm-ssh connect <vm_name>
```

If you need to specify a different user or bridge:

```bash
$ kvm-ssh connect --user <ssh_username> --bridge <bridge_name> <vm_name>
```

### Forward a port from a KVM virtual machine to the local machine

```bash
$ kvm-ssh forward --local-port <local_port> --remote-port <remote_port> <vm_name>
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
