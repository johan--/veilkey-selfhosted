package api

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/*
var embeddedAssets embed.FS

func (s *Server) assetHandler() http.Handler {
	if devDir := strings.TrimSpace(os.Getenv("VEILKEY_UI_DEV_DIR")); devDir != "" {
		assetsDir := filepath.Join(devDir, "assets")
		if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
			return http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
		}
	}
	sub, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		panic(err)
	}
	return http.StripPrefix("/assets/", http.FileServer(http.FS(sub)))
}
