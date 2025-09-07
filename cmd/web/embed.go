package web

import "embed"

//go:generate go tool templ generate ./...
//go:generate npx @tailwindcss/cli -i assets/css/input.css -o static/css/output.css --minify
//go:generate go tool esbuild --bundle --minify assets/js/app.js --outdir=static/js

//go:embed static/*
var Static embed.FS
