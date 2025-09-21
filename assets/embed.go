package assets

import (
	"crypto/sha256"
	"embed"
	"fmt"
)

//go:generate npx @tailwindcss/cli -i assets/css/input.css -o assets/build/css/app.css --cwd .. --minify
//go:generate npx esbuild --bundle --minify js/app.js --outdir=build/js

//go:embed build/*
var Static embed.FS

//go:embed robots.txt
var RobotsTxt string

var AssetVersions = func() map[string]string {
	versions := make(map[string]string)

	files := []string{
		"build/css/app.css",
		"build/js/app.js",
	}

	for _, file := range files {
		if data, err := Static.ReadFile(file); err == nil {
			hash := sha256.Sum256(data)
			versions[file] = fmt.Sprintf("%x", hash)[:8]
		}
	}

	return versions
}()
