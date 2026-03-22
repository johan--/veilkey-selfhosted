package commands

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"veilkey-vaultcenter/internal/tui"
)

func RunKeycenter() {
	url := os.Getenv("VEILKEY_KEYCENTER_URL")
	if url == "" {
		url = os.Getenv("VEILKEY_ADDR")
	}
	if url == "" {
		url = "https://10.50.0.110:11181"
	}
	// Ensure URL has scheme
	if len(url) > 0 && url[0] == ':' {
		url = "https://localhost" + url
	} else if len(url) > 0 && url[0] != 'h' {
		url = "https://" + url
	}

	fmt.Printf("Connecting to VaultCenter at %s...\n", url)

	p := tea.NewProgram(
		tui.NewModel(url),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
