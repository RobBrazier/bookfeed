package web

import (
	"crypto/sha256"
	"embed"
	"fmt"
)

//go:generate go tool templ generate ./...
//go:generate npx @tailwindcss/cli -i assets/css/input.css -o static/css/output.css --minify
//go:generate go tool esbuild --bundle --minify assets/js/app.js --outdir=static/js

//go:embed static/*
var Static embed.FS

var AssetVersions = func() map[string]string {
	versions := make(map[string]string)
	
	files := []string{
		"static/css/output.css",
		"static/js/app.js",
	}
	
	for _, file := range files {
		if data, err := Static.ReadFile(file); err == nil {
			hash := sha256.Sum256(data)
			versions[file] = fmt.Sprintf("%x", hash)[:8]
		}
	}
	
	return versions
}()
