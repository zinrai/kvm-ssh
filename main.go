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
	user   string
	bridge string
	ports  []string
)

func checkCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func getVMList() ([]string, error) {
	if !checkCommand("sudo") {
		return nil, fmt.Errorf("sudo command not found")
	}
	if !checkCommand("virsh") {
		return nil, fmt.Errorf("virsh command not found")
	}

	cmd := exec.Command("sudo", "virsh", "list", "--name", "--state-running")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute virsh command: %v\nOutput: %s", err, output)
	}

	vmNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	var filteredVMNames []string
	for _, name := range vmNames {
		if name != "" {
			filteredVMNames = append(filteredVMNames, name)
		}
	}

	return filteredVMNames, nil
}

func getVMIP(vmName, bridgeName string) (string, error) {
	statusFile := fmt.Sprintf("/var/lib/libvirt/dnsmasq/%s.status", bridgeName)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return "", fmt.Errorf("failed to read status file: %v", err)
	}

	var vmInfos []VMInfo
	if err := json.Unmarshal(data, &vmInfos); err != nil {
		return "", fmt.Errorf("failed to parse status file: %v", err)
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

	if !checkCommand("ssh") {
		return fmt.Errorf("ssh command not found")
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
		vms, err := getVMList()
		if err != nil {
			return err
		}
		for _, vm := range vms {
			fmt.Println(vm)
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
		return sshToVM(vmName, user, bridge, nil, false)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward [vm_name]",
	Short: "Forward ports from a KVM virtual machine to the local machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		return sshToVM(vmName, user, bridge, ports, true)
	},
}

func init() {
	// Set default values
	bridge = "virbr0"
	user = os.Getenv("USER")

	connectCmd.Flags().StringVarP(&user, "user", "u", user, "SSH user")
	connectCmd.Flags().StringVarP(&bridge, "bridge", "b", bridge, "Bridge name")

	forwardCmd.Flags().StringVarP(&user, "user", "u", user, "SSH user")
	forwardCmd.Flags().StringVarP(&bridge, "bridge", "b", bridge, "Bridge name")
	forwardCmd.Flags().StringSliceVarP(&ports, "port", "p", []string{}, "Ports to forward (comma-separated)")
	forwardCmd.MarkFlagRequired("port")

	rootCmd.AddCommand(listCmd, connectCmd, forwardCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
