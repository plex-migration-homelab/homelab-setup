//go:build lint
// +build lint

package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// promptReader returns a buffered reader for stdin.
func promptReader() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

func readLine(prompt string) (string, error) {
	fmt.Printf("%s ", prompt)
	line, err := promptReader().ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// PromptYesNo provides a minimal interactive prompt without external deps when lint build tag is set.
func (u *UI) PromptYesNo(prompt string, defaultYes bool) (bool, error) {
	if u.nonInteractive {
		u.Infof("[Non-interactive] %s -> %v (default)", prompt, defaultYes)
		return defaultYes, nil
	}

	for {
		def := "y"
		if !defaultYes {
			def = "n"
		}
		answer, err := readLine(fmt.Sprintf("%s [%s/%s]", prompt, strings.ToUpper(def), opposite(def)))
		if err != nil {
			return false, err
		}

		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "" {
			return defaultYes, nil
		}
		if answer == "y" || answer == "yes" {
			return true, nil
		}
		if answer == "n" || answer == "no" {
			return false, nil
		}

		u.Warning("Please enter y or n")
	}
}

func opposite(v string) string {
	if strings.ToLower(v) == "y" {
		return "n"
	}
	return "y"
}

// PromptInput returns free-form text input.
func (u *UI) PromptInput(prompt, defaultValue string) (string, error) {
	if u.nonInteractive {
		if defaultValue == "" {
			return "", fmt.Errorf("non-interactive mode requires a default value for: %s", prompt)
		}
		u.Infof("[Non-interactive] %s -> %s (default)", prompt, defaultValue)
		return defaultValue, nil
	}

	line, err := readLine(fmt.Sprintf("%s (default: %s)", prompt, defaultValue))
	if err != nil {
		return "", err
	}
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

// PromptPassword is not available in lint builds; return error to avoid unsafe echoing.
func (u *UI) PromptPassword(prompt string) (string, error) {
	return "", errors.New("password prompts are unavailable in lint build")
}

func (u *UI) PromptPasswordConfirm(prompt string) (string, error) {
	return "", errors.New("password confirmation is unavailable in lint build")
}

func (u *UI) PromptSelect(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options provided")
	}

	if u.nonInteractive {
		u.Infof("[Non-interactive] %s -> %s (first option)", prompt, options[0])
		return 0, nil
	}

	u.Printf("%s\n", prompt)
	for i, opt := range options {
		u.Printf("  %d) %s\n", i+1, opt)
	}

	for {
		line, err := readLine("Enter number")
		if err != nil {
			return -1, err
		}
		idx, err := strconv.Atoi(line)
		if err != nil || idx < 1 || idx > len(options) {
			u.Warning("Enter a number from the list")
			continue
		}
		return idx - 1, nil
	}
}

func (u *UI) PromptMultiSelect(prompt string, options []string) ([]int, error) {
	if len(options) == 0 {
		return []int{}, nil
	}

	if u.nonInteractive {
		indices := make([]int, len(options))
		for i := range options {
			indices[i] = i
		}
		u.Infof("[Non-interactive] %s -> all options (%d)", prompt, len(options))
		return indices, nil
	}

	u.Printf("%s\n", prompt)
	for i, opt := range options {
		u.Printf("  %d) %s\n", i+1, opt)
	}

	line, err := readLine("Enter comma-separated numbers (leave blank for none)")
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return []int{}, nil
	}

	parts := strings.Split(line, ",")
	var indices []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx, err := strconv.Atoi(part)
		if err != nil || idx < 1 || idx > len(options) {
			return nil, fmt.Errorf("invalid selection: %s", part)
		}
		indices = append(indices, idx-1)
	}
	return indices, nil
}

func (u *UI) PromptInputRequired(prompt string) (string, error) {
	for {
		value, err := u.PromptInput(prompt, "")
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) == "" {
			u.Warning("Value cannot be empty")
			continue
		}
		return value, nil
	}
}

func (u *UI) PromptInputWithValidation(prompt, defaultValue string, _ interface{}) (string, error) {
	// Validation is skipped in lint builds; rely on subsequent checks.
	return u.PromptInput(prompt, defaultValue)
}
