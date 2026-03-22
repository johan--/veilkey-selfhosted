package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type loginStep int

const (
	loginStepCheckStatus loginStep = iota
	loginStepUnlock
	loginStepTOTP
)

type loginModel struct {
	step       loginStep
	kekInput   textinput.Model
	totpInput  textinput.Model
	logging    bool
	errText    string
	serverLocked bool
}

type unlockSuccessMsg struct{}
type unlockFailMsg struct{ err string }

func newLoginModel() loginModel {
	kek := textinput.New()
	kek.Placeholder = "master password (KEK)"
	kek.EchoMode = textinput.EchoPassword
	kek.EchoCharacter = '•'
	kek.Width = 40

	totp := textinput.New()
	totp.Placeholder = "6-digit TOTP code"
	totp.CharLimit = 6
	totp.Width = 20

	return loginModel{
		step:      loginStepCheckStatus,
		kekInput:  kek,
		totpInput: totp,
	}
}

func unlockCmd(c *Client, password string) tea.Cmd {
	return func() tea.Msg {
		if err := c.Unlock(password); err != nil {
			return unlockFailMsg{err.Error()}
		}
		return unlockSuccessMsg{}
	}
}

func loginCmd(c *Client, code string) tea.Cmd {
	return func() tea.Msg {
		if err := c.Login(code); err != nil {
			return loginFailMsg{err.Error()}
		}
		return loginSuccessMsg{}
	}
}

func (m loginModel) update(msg tea.Msg, c *Client) (loginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		if msg.status == "locked" {
			m.step = loginStepUnlock
			m.serverLocked = true
			m.kekInput.Focus()
		} else if msg.status == "ready" {
			m.step = loginStepTOTP
			m.totpInput.Focus()
		} else {
			m.errText = "Server is " + msg.status
		}
		return m, nil

	case unlockSuccessMsg:
		m.step = loginStepTOTP
		m.logging = false
		m.errText = ""
		m.totpInput.SetValue("")
		m.totpInput.Focus()
		return m, nil

	case unlockFailMsg:
		m.logging = false
		m.errText = msg.err
		m.kekInput.SetValue("")
		m.kekInput.Focus()
		return m, nil

	case loginFailMsg:
		m.logging = false
		m.errText = msg.err
		m.totpInput.SetValue("")
		m.totpInput.Focus()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.step {
			case loginStepUnlock:
				pw := strings.TrimSpace(m.kekInput.Value())
				if pw == "" {
					return m, nil
				}
				m.logging = true
				m.errText = ""
				return m, unlockCmd(c, pw)
			case loginStepTOTP:
				code := strings.TrimSpace(m.totpInput.Value())
				if code == "" {
					return m, nil
				}
				m.logging = true
				m.errText = ""
				return m, loginCmd(c, code)
			}
		}
	}

	// Update active input
	var cmd tea.Cmd
	switch m.step {
	case loginStepUnlock:
		m.kekInput, cmd = m.kekInput.Update(msg)
	case loginStepTOTP:
		m.totpInput, cmd = m.totpInput.Update(msg)
	}
	return m, cmd
}

func (m loginModel) view(width int) string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("🔐 VeilKey VaultCenter"))
	b.WriteString("\n\n")

	if m.step == loginStepCheckStatus {
		b.WriteString("  " + styleDim.Render("Connecting..."))
		return b.String()
	}

	if m.logging {
		if m.step == loginStepUnlock {
			b.WriteString("  " + styleDim.Render("Unlocking server..."))
		} else {
			b.WriteString("  " + styleDim.Render("Authenticating..."))
		}
		return b.String()
	}

	switch m.step {
	case loginStepUnlock:
		b.WriteString("  " + styleError.Render("Server is locked") + "\n\n")
		b.WriteString("  " + styleLabel.Render("Master Key") + "\n")
		b.WriteString("  " + m.kekInput.View() + "\n\n")
	case loginStepTOTP:
		if m.serverLocked {
			b.WriteString("  " + lipglossGreen("✓ Server unlocked") + "\n\n")
		}
		b.WriteString("  " + styleLabel.Render("TOTP Code") + "\n")
		b.WriteString("  " + m.totpInput.View() + "\n\n")
	}

	if m.errText != "" {
		b.WriteString("  " + styleError.Render(m.errText) + "\n\n")
	}

	b.WriteString(styleDim.Render("  enter submit  q quit"))

	return b.String()
}

func lipglossGreen(s string) string {
	return styleSuccess.Render(s)
}
