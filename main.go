package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type VMInfo struct {
	IPAddress  string `json:"ip-address"`
	MacAddress string `json:"mac-address"`
	Hostname   string `json:"hostname"`
	ClientID   string `json:"client-id"`
	ExpiryTime int64  `json:"expiry-time"`
}

var (
	user  string
	ports []string
)

func getVMList(bridgeName string) ([]VMInfo, error) {
	statusFile := fmt.Sprintf("/var/lib/libvirt/dnsmasq/%s.status", bridgeName)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file: %v", err)
	}

	var vmInfos []VMInfo
	if err := json.Unmarshal(data, &vmInfos); err != nil {
		return nil, fmt.Errorf("failed to parse status file: %v", err)
	}

	return vmInfos, nil
}

func getVMIP(vmName, bridgeName string) (string, error) {
	vmInfos, err := getVMList(bridgeName)
	if err != nil {
		return "", err
	}

	for _, info := range vmInfos {
		if info.Hostname == vmName {
			return info.IPAddress, nil
		}
	}

	return "", fmt.Errorf("VM not found: %s", vmName)
}

func sshToVM(vmName, user, bridgeName string, ports []string, isForward bool) error {
	ip, err := getVMIP(vmName, bridgeName)
	if err != nil {
		return err
	}

	var sshArgs []string
	if isForward {
		for _, port := range ports {
			sshArgs = append(sshArgs, "-L", fmt.Sprintf("localhost:%s:localhost:%s", port, port))
		}
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, ip))

	fmt.Printf("Executing: ssh %s\n", strings.Join(sshArgs, " "))

	sshCmd := exec.Command("ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

var rootCmd = &cobra.Command{
	Use:   "kvm-ssh",
	Short: "SSH into KVM virtual machines and perform port forwarding",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all running KVM virtual machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		bridge, _ := cmd.Flags().GetString("bridge")
		vmInfos, err := getVMList(bridge)
		if err != nil {
			return err
		}
		for _, vm := range vmInfos {
			fmt.Printf("Hostname: %s, IP: %s\n", vm.Hostname, vm.IPAddress)
		}
		return nil
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect [vm_name]",
	Short: "Connect to a KVM virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		bridge, _ := cmd.Flags().GetString("bridge")
		user, _ := cmd.Flags().GetString("user")
		return sshToVM(vmName, user, bridge, nil, false)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward [vm_name]",
	Short: "Forward ports from a KVM virtual machine to the local machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		bridge, _ := cmd.Flags().GetString("bridge")
		user, _ := cmd.Flags().GetString("user")
		ports, _ := cmd.Flags().GetStringSlice("port")
		return sshToVM(vmName, user, bridge, ports, true)
	},
}

func init() {
	defaultBridge := "virbr0"
	user = os.Getenv("USER")

	listCmd.Flags().StringP("bridge", "b", defaultBridge, "Bridge name")

	connectCmd.Flags().StringP("bridge", "b", defaultBridge, "Bridge name")
	connectCmd.Flags().StringP("user", "u", user, "SSH user")

	forwardCmd.Flags().StringP("bridge", "b", defaultBridge, "Bridge name")
	forwardCmd.Flags().StringP("user", "u", user, "SSH user")
	forwardCmd.Flags().StringSliceP("port", "p", []string{}, "Ports to forward (comma-separated)")
	forwardCmd.MarkFlagRequired("port")

	rootCmd.AddCommand(listCmd, connectCmd, forwardCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
