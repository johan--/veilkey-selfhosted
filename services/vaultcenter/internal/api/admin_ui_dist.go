package api

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed ui_dist/* ui_dist/assets/*
var adminUIDist embed.FS

func uiDevDir() string {
	return strings.TrimSpace(os.Getenv("VEILKEY_UI_DEV_DIR"))
}

func devUIIndex() ([]byte, bool) {
	devDir := uiDevDir()
	if devDir == "" {
		return nil, false
	}
	path := filepath.Join(devDir, "index.html")
	if body, err := os.ReadFile(path); err == nil {
		return body, true
	}
	return nil, false
}

func embeddedUIIndex() ([]byte, bool) {
	body, err := fs.ReadFile(adminUIDist, "ui_dist/index.html")
	if err != nil {
		return nil, false
	}
	return body, true
}

func devUIStaticFile(name string) ([]byte, bool) {
	devDir := uiDevDir()
	if devDir == "" {
		return nil, false
	}
	path := filepath.Join(devDir, filepath.Clean(name))
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return body, true
}

func embeddedUIStaticFile(name string) ([]byte, bool) {
	clean := strings.TrimPrefix(filepath.Clean(name), "/")
	if clean == "" || strings.HasPrefix(clean, "assets/") {
		return nil, false
	}
	body, err := fs.ReadFile(adminUIDist, filepath.Join("ui_dist", clean))
	if err != nil {
		return nil, false
	}
	return body, true
}

func devUIAssetsDir() string {
	devDir := uiDevDir()
	if devDir == "" {
		return ""
	}
	assetsDir := filepath.Join(devDir, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		return assetsDir
	}
	return ""
}

func embeddedUIAssets() (fs.FS, bool) {
	sub, err := fs.Sub(adminUIDist, "ui_dist/assets")
	if err != nil {
		return nil, false
	}
	return sub, true
}
