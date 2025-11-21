package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/zoro11031/homelab-coreos-minipc/homelab-setup/internal/troubleshoot"
)

// ErrExit is returned when the user chooses to exit the menu
var ErrExit = errors.New("exit")

// ErrBack is returned when the user chooses to go back
var ErrBack = errors.New("back")

// Menu provides an interactive menu interface
type Menu struct {
	ctx *SetupContext
}

// NewMenu creates a new Menu instance
func NewMenu(ctx *SetupContext) *Menu {
	return &Menu{ctx: ctx}
}

// clearScreen clears the terminal screen using ANSI escape codes
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// Show displays the main menu and handles user input
func (m *Menu) Show() error {
	for {
		clearScreen()
		m.displayMenu()

		choice, err := m.ctx.UI.PromptInput("Enter your choice", "")
		if err != nil {
			return err
		}

		choice = strings.ToUpper(strings.TrimSpace(choice))

		if err := m.handleChoice(choice); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			if errors.Is(err, ErrBack) {
				continue
			}
			m.ctx.UI.Error(fmt.Sprintf("%v", err))
			m.ctx.UI.Print("")
			m.ctx.UI.Info("Press Enter to continue...")
			_, _ = fmt.Scanln()
		}
	}
}

// displayMenu displays the main menu
func (m *Menu) displayMenu() {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	// Header
	border := strings.Repeat("=", 70)
	cyan.Println(border)
	cyan.Println("  UBlue uCore Homelab Setup & Management")
	cyan.Println(border)
	fmt.Println()

	m.ctx.UI.Info("Select an option:")
	fmt.Println()

	bold.Println("  [1] NFS Management Tool")
	bold.Println("  [2] WireGuard Management Tool")
	bold.Println("  [3] Network Troubleshooting Suite")
	bold.Println("  [4] Factory Reset / Legacy Setup")
	fmt.Println()
	bold.Println("  [H] Help")
	bold.Println("  [X] Exit")
	fmt.Println()
}

// handleChoice processes the user's menu choice
func (m *Menu) handleChoice(choice string) error {
	switch choice {
	case "1":
		return m.runNFSManagement()
	case "2":
		return m.runWireGuardManagement()
	case "3":
		return m.runTroubleshooting()
	case "4":
		return m.runLegacySetup()
	case "H":
		return m.showHelp()
	case "X":
		return ErrExit
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

// runNFSManagement runs the NFS setup step
func (m *Menu) runNFSManagement() error {
	// Currently just links to the existing NFS setup step
	// Could be expanded to a submenu in the future
	return m.runStepWrapper("nfs", "NFS Setup")
}

// runWireGuardManagement runs the WireGuard management submenu
func (m *Menu) runWireGuardManagement() error {
	for {
		clearScreen()
		m.ctx.UI.Header("WireGuard Management")

		fmt.Println("  [1] Run WireGuard Setup (Configure Server)")
		fmt.Println("  [2] Add WireGuard Peer")
		fmt.Println("  [B] Back")
		fmt.Println()

		choice, err := m.ctx.UI.PromptInput("Enter your choice", "")
		if err != nil {
			return err
		}

		switch strings.ToUpper(strings.TrimSpace(choice)) {
		case "1":
			if err := m.runStepWrapper("wireguard", "WireGuard Setup"); err != nil {
				m.ctx.UI.Error(err.Error())
				m.waitEnter()
			}
		case "2":
			clearScreen()
			m.ctx.UI.Header("Add WireGuard Peer")
			if err := AddWireGuardPeer(m.ctx, nil); err != nil {
				m.ctx.UI.Error(err.Error())
			}
			m.waitEnter()
		case "B":
			return nil
		default:
			m.ctx.UI.Error("Invalid choice")
			m.waitEnter()
		}
	}
}

// runTroubleshooting runs the troubleshooting tool
func (m *Menu) runTroubleshooting() error {
	clearScreen()
	if err := troubleshoot.Run(m.ctx.Config, m.ctx.UI); err != nil {
		return err
	}
	m.waitEnter()
	return nil
}

// runLegacySetup runs the legacy setup menu with triple confirmation
func (m *Menu) runLegacySetup() error {
	clearScreen()
	m.ctx.UI.Header("Factory Reset / Legacy Setup")
	m.ctx.UI.Warning("DANGER: This menu contains legacy setup steps that may overwrite configuration.")
	m.ctx.UI.Warning("These steps are intended for initial factory setup or full resets.")
	fmt.Println()

	// Triple confirmation
	for i := 1; i <= 3; i++ {
		confirm, err := m.ctx.UI.PromptYesNo(fmt.Sprintf("Are you sure you want to proceed? (%d/3)", i), false)
		if err != nil {
			return err
		}
		if !confirm {
			m.ctx.UI.Info("Operation cancelled.")
			m.waitEnter()
			return nil
		}
	}

	return m.showLegacyMenu()
}

// showLegacyMenu displays the legacy setup menu
func (m *Menu) showLegacyMenu() error {
	for {
		clearScreen()
		m.ctx.UI.Header("Legacy Setup Menu")
		m.ctx.UI.Warning("Warning: These steps modify system configuration.")
		fmt.Println()

		fmt.Println("  [1] Run All Legacy Steps")
		fmt.Println("  [2] User Setup")
		fmt.Println("  [3] Directory Setup")
		fmt.Println("  [4] Container Setup")
		fmt.Println("  [5] Service Deployment")
		fmt.Println("  [6] Reset Setup Markers")
		fmt.Println("  [7] Show Setup Status")
		fmt.Println("  [B] Back to Main Menu")
		fmt.Println()

		choice, err := m.ctx.UI.PromptInput("Enter your choice", "")
		if err != nil {
			return err
		}

		var errOp error
		switch strings.ToUpper(strings.TrimSpace(choice)) {
		case "1":
			errOp = m.runAllLegacySteps()
		case "2":
			errOp = m.runStepWrapper("user", "User Setup")
		case "3":
			errOp = m.runStepWrapper("directory", "Directory Setup")
		case "4":
			errOp = m.runStepWrapper("container", "Container Setup")
		case "5":
			errOp = m.runStepWrapper("deployment", "Service Deployment")
		case "6":
			errOp = m.resetSetup()
		case "7":
			errOp = m.showStatus()
		case "B":
			return nil
		default:
			m.ctx.UI.Error("Invalid choice")
			m.waitEnter()
			continue
		}

		if errOp != nil {
			m.ctx.UI.Error(errOp.Error())
			m.waitEnter()
		}
	}
}

// runStepWrapper runs a single setup step with header and pause
func (m *Menu) runStepWrapper(stepName, displayName string) error {
	clearScreen()
	m.ctx.UI.Header(displayName)

	err := RunStep(m.ctx, stepName)

	fmt.Println()
	m.ctx.UI.Info("Press Enter to return to menu...")
	_, _ = fmt.Scanln()

	return err
}

// runAllLegacySteps runs all legacy steps in order
func (m *Menu) runAllLegacySteps() error {
	clearScreen()
	m.ctx.UI.Header("Running Complete Legacy Setup")
	m.ctx.UI.Info("Note: WireGuard is skipped in this flow (use Main Menu > WireGuard Management)")

	// RunAll handles: preflight, user, directory, (wireguard skipped), nfs, container, deployment
	// We pass true to skipWireGuard
	err := RunAll(m.ctx, true)

	m.waitEnter()
	return err
}

// showStatus shows the current setup status
func (m *Menu) showStatus() error {
	clearScreen()
	m.ctx.UI.Header("Setup Status")

	fmt.Println()
	m.ctx.UI.Info("Completed Steps:")
	fmt.Println()

	steps := GetAllSteps()
	completedCount := 0

	for i, step := range steps {
		if IsStepComplete(m.ctx.Config, step.MarkerName) {
			m.ctx.UI.Successf("[%d] ✓ %s", i, step.Name)
			completedCount++
		} else {
			m.ctx.UI.Infof("[%d] - %s (not completed)", i, step.Name)
		}
	}

	fmt.Println()
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Println(strings.Repeat("-", 70))
	m.ctx.UI.Infof("Progress: %d/%d steps completed", completedCount, len(steps))
	cyan.Println(strings.Repeat("-", 70))
	fmt.Println()

	// Show configuration file location
	if _, err := os.Stat(m.ctx.Config.FilePath()); err == nil {
		m.ctx.UI.Infof("Configuration file: %s", m.ctx.Config.FilePath())
	}

	m.waitEnter()
	return nil
}

// resetSetup resets all completion markers
func (m *Menu) resetSetup() error {
	clearScreen()
	m.ctx.UI.Header("Reset Setup")

	m.ctx.UI.Warning("This will clear all completion markers")
	m.ctx.UI.Warning("Configuration file will NOT be deleted")
	fmt.Println()

	confirm, err := m.ctx.UI.PromptYesNo("Are you sure you want to reset markers?", false)
	if err != nil {
		return err
	}

	if !confirm {
		m.ctx.UI.Info("Reset cancelled")
		m.waitEnter()
		return nil
	}

	if err := m.ctx.Config.ClearAllMarkers(); err != nil {
		return fmt.Errorf("failed to remove markers: %w", err)
	}

	m.ctx.UI.Success("All completion markers have been cleared")
	m.waitEnter()

	return nil
}

func (m *Menu) waitEnter() {
	fmt.Println()
	m.ctx.UI.Info("Press Enter to return to menu...")
	_, _ = fmt.Scanln()
}

// showHelp displays help information
func (m *Menu) showHelp() error {
	clearScreen()
	m.ctx.UI.Header("Help")

	help := `
UBlue uCore Homelab Setup & Management

MAIN OPTIONS:

  1. NFS Management Tool
     - Configure NFS client connections and mounts.

  2. WireGuard Management Tool
     - Configure WireGuard VPN server.
     - Add new peers (clients).

  3. Network Troubleshooting Suite
     - Ping check (File Server).
     - DNS diagnostics.
     - Port scanning (VPS/Portainer).

  4. Factory Reset / Legacy Setup (Advanced)
     - Access original setup steps (User, Directory, Container, Deployment).
     - Use this for initial setup or resetting the environment.
     - Requires triple confirmation.

DOCUMENTATION:
  For more information, see the project documentation or README.
`

	fmt.Println(help)
	m.waitEnter()

	return nil
}
