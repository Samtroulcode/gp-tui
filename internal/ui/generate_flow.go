package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gopass-tui/internal/gopass"
)

const defaultQuickPasswordLength = 24

type generateStep int

const (
	generateStepOverwriteConfirm generateStep = iota
	generateStepQuickConfirm
	generateStepKey
	generateStepGenerator
	generateStepLength
	generateStepSymbols
	generateStepStrict
	generateStepSeparator
	generateStepLanguage
)

type generationFlow struct {
	creatingNew bool
	request     gopass.GenerateRequest
	step        generateStep
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
	case "n", "N":
		return m.submitGenerateWizardConfirm(false)
	case "enter":
		return m.submitGenerateWizardConfirm(m.defaultConfirmAnswer())
	}

	return nil
}

func (m *Model) defaultConfirmAnswer() bool {
	if m.input.generation != nil && m.input.generation.step == generateStepQuickConfirm {
		return true
	}

	return false
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
		if flow.creatingNew && !answer {
			entryPath := flow.request.Path
			m.input = inputState{}
			m.setStatus("creating entry %s", entryPath)
			return createEntryCmd(m.service, entryPath)
		}
		if !flow.creatingNew && !answer {
			m.input = inputState{}
			m.setStatus("cancelled")
			return nil
		}
		if !flow.creatingNew {
			flow.request.Force = true
		}
		m.showGenerateStep(flow, generateStepQuickConfirm)
		return nil

	case generateStepQuickConfirm:
		if answer {
			flow.request = quickGenerateRequest(flow.request.Path, flow.request.Force)
			return m.completeGenerateWizard(flow)
		}
		m.showGenerateStep(flow, generateStepKey)
		return nil

	case generateStepSymbols:
		flow.request.Symbols = answer
	case generateStepStrict:
		flow.request.Strict = answer
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
	case generateStepGenerator:
		generator := strings.ToLower(value)
		switch generator {
		case "cryptic", "memorable", "xkcd", "external":
			flow.request.Generator = generator
		case "":
			flow.request.Generator = "cryptic"
		default:
			m.setStatus("generator must be cryptic, memorable, xkcd, or external")
			return nil
		}
	case generateStepLength:
		length, err := strconv.Atoi(value)
		if err != nil || length <= 0 {
			m.setStatus("password length must be a positive number")
			return nil
		}
		flow.request.Length = length
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
	default:
		return nil
	}

	return m.showNextGenerateStep(flow)
}

func (m *Model) completeGenerateWizard(flow *generationFlow) tea.Cmd {
	request := flow.request
	m.input = inputState{}
	m.setStatus("generating password for %s", request.Path)
	return generateEntryCmd(m.service, request, flow.creatingNew)
}

func (m *Model) beginGenerateFlow(entryPath string, creatingNew bool) {
	request := defaultGenerateRequest(entryPath)

	flow := &generationFlow{
		creatingNew: creatingNew,
		request:     request,
		step:        generateStepOverwriteConfirm,
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
	case generateStepQuickConfirm:
		input.promptKind = inputPromptConfirm
		input.prompt = "Quick generation with recommended defaults? [Y/n]"
	case generateStepKey:
		input.prompt = "Secret key (blank for password line)"
		input.value = flow.request.Key
	case generateStepGenerator:
		input.prompt = "Generator [cryptic|memorable|xkcd|external]"
		input.value = flow.request.Generator
	case generateStepLength:
		input.prompt = "Password length"
		input.value = strconv.Itoa(flow.request.Length)
	case generateStepSymbols:
		input.prompt = "Use symbols? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepStrict:
		input.prompt = "Require strict character rules? [y/N]"
		input.promptKind = inputPromptConfirm
	case generateStepSeparator:
		input.prompt = "Word separator (optional)"
		input.value = flow.request.Separator
	case generateStepLanguage:
		input.prompt = "Language [en|de]"
		input.value = flow.request.Language
	}

	m.input = input
}

func (flow *generationFlow) nextStep() (generateStep, bool) {
	switch flow.step {
	case generateStepOverwriteConfirm:
		return generateStepQuickConfirm, false
	case generateStepQuickConfirm:
		return generateStepKey, false
	case generateStepKey:
		return generateStepGenerator, false
	case generateStepGenerator:
		return generateStepLength, false
	case generateStepLength:
		switch flow.request.Generator {
		case "cryptic":
			return generateStepSymbols, false
		case "xkcd":
			return generateStepSeparator, false
		default:
			return 0, true
		}
	case generateStepSymbols:
		return generateStepStrict, false
	case generateStepStrict:
		return 0, true
	case generateStepSeparator:
		return generateStepLanguage, false
	case generateStepLanguage:
		return 0, true
	default:
		return 0, true
	}
}

func quickGenerateRequest(path string, force bool) gopass.GenerateRequest {
	request := defaultGenerateRequest(path)
	request.Force = force
	request.Symbols = true
	request.Strict = true
	return request
}

func defaultGenerateRequest(path string) gopass.GenerateRequest {
	return gopass.GenerateRequest{
		Path:      path,
		Length:    defaultQuickPasswordLength,
		Generator: "cryptic",
		Language:  "en",
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
