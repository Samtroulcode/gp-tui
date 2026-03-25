package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
)

func (m *Model) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch m.input.mode {
	case inputModeDeleteEntries:
		return m.handleDeleteConfirmInput(msg)
	case inputModeSearch:
		return m.handleSearchInput(msg)
	case inputModeGenerateWizard:
		return m.handleGenerateWizardInput(msg)
	default:
		return m.handleTextPromptInput(msg)
	}
}

func (m *Model) handleTextPromptInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "enter":
		return m.submitInput()
	case "backspace":
		if len(m.input.value) == 0 {
			return nil
		}

		runes := []rune(m.input.value)
		m.input.value = string(runes[:len(runes)-1])
		return nil
	}

	if len(msg.Runes) > 0 {
		m.input.value += string(msg.Runes)
	}

	return nil
}

func (m *Model) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.finishSearch(true)
		m.setStatus("cancelled")
		return nil
	case "enter":
		m.searchQuery = strings.TrimSpace(m.input.value)
		if m.searchQuery == "" {
			m.finishSearch(true)
			return nil
		}
		if m.currentNode() == nil {
			m.finishSearch(true)
			m.setStatus("no matching entry selected")
			return nil
		}

		m.finishSearchWithSelection()
		return nil
	case "backspace":
		if len(m.input.value) == 0 {
			return nil
		}

		runes := []rune(m.input.value)
		m.input.value = string(runes[:len(runes)-1])
		m.searchQuery = m.input.value
		m.applySearchFilter()
		return nil
	}

	if len(msg.Runes) > 0 {
		m.input.value += string(msg.Runes)
		m.searchQuery = m.input.value
		m.applySearchFilter()
	}

	return nil
}

func (m *Model) handleDeleteConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "y", "enter":
		return m.submitInput()
	}

	return nil
}

func (m *Model) handleGenerateWizardInput(msg tea.KeyMsg) tea.Cmd {
	if m.input.promptKind == inputPromptConfirm {
		return m.handleGenerateWizardConfirmInput(msg)
	}

	return m.handleTextPromptInput(msg)
}

func (m *Model) handleGenerateWizardConfirmInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	case "y", "Y":
		return m.submitGenerateWizardConfirm(true)
	case "n", "N", "enter":
		return m.submitGenerateWizardConfirm(false)
	}

	return nil
}

func (m *Model) submitInput() tea.Cmd {
	if m.input.mode == inputModeDeleteEntries {
		paths := append([]string(nil), m.input.paths...)
		m.input = inputState{}
		m.setStatus("deleting %s", entryCountLabel(len(paths)))

		focusPath := m.currentDirectory()
		expanded := m.expandedStateForReload()
		if focusPath != "" {
			expanded[focusPath] = true
		}

		return deleteEntriesCmd(m.service, paths, focusPath, expanded)
	}

	switch m.input.mode {
	case inputModeCreateEntry:
		return m.submitCreateEntryPath()
	case inputModeGenerateWizard:
		return m.submitGenerateWizardText()
	default:
		m.input = inputState{}
		return nil
	}
}

func (m *Model) submitCreateEntryPath() tea.Cmd {
	entryPath := strings.Trim(strings.TrimSpace(m.input.value), "/")
	if entryPath == "" {
		m.setStatus("entry path is required")
		return nil
	}

	m.beginGenerateFlow(entryPath, true)
	return nil
}

func (m *Model) submitGenerateWizardConfirm(answer bool) tea.Cmd {
	flow := m.input.generation
	if flow == nil {
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	}

	switch flow.step {
	case generateStepOverwriteConfirm:
		if flow.creatingNew {
			if !answer {
				entryPath := flow.request.Path
				m.input = inputState{}
				m.setStatus("creating entry %s", entryPath)
				return createEntryCmd(m.service, entryPath)
			}

			m.showGenerateStep(flow, generateStepKey)
			return nil
		}

		if !answer {
			m.input = inputState{}
			m.setStatus("cancelled")
			return nil
		}

		flow.request.Force = true
		m.showGenerateStep(flow, generateStepKey)
		return nil

	case generateStepSymbols:
		flow.request.Symbols = answer
	case generateStepStrict:
		flow.request.Strict = answer
	case generateStepForceRegen:
		flow.request.ForceRegen = answer
	case generateStepClip:
		flow.request.Clip = answer
	case generateStepPrint:
		flow.request.Print = answer
	case generateStepEdit:
		flow.request.Edit = answer
	case generateStepInteractiveCommit:
		flow.request.InteractiveCommit = answer
		return m.completeGenerateWizard(flow)
	default:
		return nil
	}

	return m.showNextGenerateStep(flow)
}

func (m *Model) submitGenerateWizardText() tea.Cmd {
	flow := m.input.generation
	if flow == nil {
		m.input = inputState{}
		m.setStatus("cancelled")
		return nil
	}

	value := strings.TrimSpace(m.input.value)

	switch flow.step {
	case generateStepKey:
		flow.request.Key = value

	case generateStepLength:
		length, err := strconv.Atoi(value)
		if err != nil || length <= 0 {
			m.setStatus("password length must be a positive number")
			return nil
		}
		flow.request.Length = length

	case generateStepGenerator:
		generator := strings.ToLower(value)
		switch generator {
		case "cryptic", "memorable", "xkcd", "external":
			flow.request.Generator = generator
		default:
			m.setStatus("generator must be cryptic, memorable, xkcd, or external")
			return nil
		}

	case generateStepSeparator:
		flow.request.Separator = value

	case generateStepLanguage:
		language := strings.ToLower(value)
		if language == "" {
			language = "en"
		}
		switch language {
		case "en", "de":
			flow.request.Language = language
		default:
			m.setStatus("language must be en or de")
			return nil
		}

	case generateStepCommitMessage:
		flow.request.CommitMessage = value

	default:
		return nil
	}

	return m.showNextGenerateStep(flow)
}

func (m *Model) completeGenerateWizard(flow *generationFlow) tea.Cmd {
	request := flow.request
	m.input = inputState{}
	m.setStatus("generating password for %s", request.Path)
	return generateEntryCmd(m.service, request)
}

func (m *Model) beginGenerateFlow(entryPath string, creatingNew bool) {
	flow := &generationFlow{
		creatingNew: creatingNew,
		request: gopass.GenerateRequest{
			Path:      entryPath,
			Length:    24,
			Generator: "cryptic",
			Language:  "en",
		},
		step: generateStepOverwriteConfirm,
	}

	if !creatingNew {
		flow.request.Force = true
	}

	m.showGenerateStep(flow, generateStepOverwriteConfirm)
}

func (m *Model) showNextGenerateStep(flow *generationFlow) tea.Cmd {
	nextStep, done := flow.nextStep()
	if done {
		return m.completeGenerateWizard(flow)
	}

	m.showGenerateStep(flow, nextStep)
	return nil
}

func (m *Model) showGenerateStep(flow *generationFlow, step generateStep) {
	flow.step = step

	input := inputState{
		mode:       inputModeGenerateWizard,
		promptKind: inputPromptText,
		generation: flow,
	}

	switch step {
	case generateStepOverwriteConfirm:
		input.promptKind = inputPromptConfirm
		if flow.creatingNew {
			input.prompt = "Generate password? [y/N]"
		} else {
			input.prompt = fmt.Sprintf("Replace the password for %s? This will overwrite the current password. [y/N]", flow.request.Path)
		}
	case generateStepKey:
		input.prompt = "Secret key (blank for password line)"
		input.value = flow.request.Key
	case generateStepLength:
		input.prompt = "Password length"
		input.value = strconv.Itoa(flow.request.Length)
	case generateStepGenerator:
		input.prompt = "Generator [cryptic|memorable|xkcd|external]"
		input.value = flow.request.Generator
	case generateStepSymbols:
		input.prompt = "Use symbols? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepStrict:
		input.prompt = "Require strict character rules? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepForceRegen:
		input.prompt = "Overwrite the entire secret? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepSeparator:
		input.prompt = "Word separator (optional)"
		input.value = flow.request.Separator
	case generateStepLanguage:
		input.prompt = "Language [en|de]"
		input.value = flow.request.Language
	case generateStepClip:
		input.prompt = "Copy generated password to the clipboard? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepPrint:
		input.prompt = "Print the generated password in the terminal? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepEdit:
		input.prompt = "Open the entry in the editor after generation? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepCommitMessage:
		input.prompt = "Commit message (optional)"
		input.value = flow.request.CommitMessage
	case generateStepInteractiveCommit:
		input.prompt = "Edit the commit message interactively? [y/N]"
		input.promptKind = inputPromptConfirm
	}

	m.input = input
}

func (flow *generationFlow) nextStep() (generateStep, bool) {
	switch flow.step {
	case generateStepOverwriteConfirm:
		return generateStepKey, false
	case generateStepKey:
		return generateStepLength, false
	case generateStepLength:
		return generateStepGenerator, false
	case generateStepGenerator:
		return generateStepSymbols, false
	case generateStepSymbols:
		return generateStepStrict, false
	case generateStepStrict:
		if flow.creatingNew {
			return generateStepSeparator, false
		}
		return generateStepForceRegen, false
	case generateStepForceRegen:
		return generateStepSeparator, false
	case generateStepSeparator:
		return generateStepLanguage, false
	case generateStepLanguage:
		return generateStepClip, false
	case generateStepClip:
		return generateStepPrint, false
	case generateStepPrint:
		return generateStepEdit, false
	case generateStepEdit:
		return generateStepCommitMessage, false
	case generateStepCommitMessage:
		return generateStepInteractiveCommit, false
	case generateStepInteractiveCommit:
		return 0, true
	default:
		return 0, true
	}
}

func (m *Model) beginCreateEntry() {
	base := m.currentDirectory()
	value := ""
	if base != "" {
		value = base + "/"
	}

	m.input = inputState{
		mode:   inputModeCreateEntry,
		prompt: "New entry",
		value:  value,
	}
}
