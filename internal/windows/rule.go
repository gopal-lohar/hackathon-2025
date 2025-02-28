package windows

import (
	"fmt"
	"os/exec"
)

func newNetShCmd(name, dir, action, program, remoteip, protocol string, enabled bool) []string {
	var actuallyEnabled string
	if enabled {
		actuallyEnabled = "yes"
	} else {
		actuallyEnabled = "no"
	}
	cmd := make([]string, 0)
	cmd = append(cmd, "advfirewall", "firewall", "add", "rule")
	cmd = append(cmd, "name="+name)
	cmd = append(cmd, "dir="+dir)
	cmd = append(cmd, "action="+action)
	cmd = append(cmd, "program="+program)
	cmd = append(cmd, "remoteip="+remoteip)
	cmd = append(cmd, "protocol="+protocol)
	cmd = append(cmd, "enable="+actuallyEnabled)
	return cmd
}

// This is required because we need to add two rules
// one for in and one for out : direction
func (w *Windows) AddNewRule(name, action, program, remoteip, protocol string) error {
	err := w.applyRule(name+"In", "in", "block", program, remoteip, protocol, true)
	if err != nil {
		return err
	}
	err = w.applyRule(name+"Out", "out", "block", program, remoteip, protocol, true)
	if err != nil {
		return err
	}
	return nil
}

// applyRule adds a firewall rule using netsh
func (w *Windows) applyRule(name, direction, action, program, remoteip, protocol string, enabled bool) error {
	// Build the netsh command
	args := newNetShCmd(name, direction, action, program, remoteip, protocol, enabled)
	// Execute the command
	w.logger.Infof("Adding rule: %v", args)
	cmd := exec.Command("netsh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add rule: %v, output: %s", err, output)
	}

	return nil
}

// RemoveRule removes a firewall rule using netsh
// Removes both In and Out rules
// Takes in the ruleName string
func (w *Windows) RemoveRule(ruleName string) error {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+ruleName+"In")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove rule: %v, output: %s", err, output)
	}
	cmd = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+ruleName+"Out")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove rule: %v, output: %s", err, output)
	}
	return nil
}

// ToggleRuleState enables or disables a firewall rule
func (w *Windows) ToggleRuleState(ruleName string, enabled bool) error {
	enabledStr := "yes"
	if !enabled {
		enabledStr = "no"
	}

	cmd := exec.Command("netsh", "advfirewall", "firewall", "set", "rule",
		"name="+ruleName, "new", "enable="+enabledStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update rule state: %v, output: %s", err, output)
	}

	return nil
}
