package assets

import (
	"embed"
)

//go:generate npx -y @tailwindcss/cli -i assets/css/input.css -o assets/build/css/app.css --cwd .. --minify
//go:generate npx -y esbuild --bundle --minify js/app.js --outdir=build/js
//go:generate npx -y esbuild --minify js/*.min.js --outdir=build/js

//go:embed build/*
var Static embed.FS

//go:embed robots.txt
var RobotsTxt string
