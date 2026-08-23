# plugin-tailwind

Tailwind CSS plugin for the [Dreego](https://github.com/dreego-stack/dreego)
framework. Integrates Tailwind CSS compilation into the `dreego build` pipeline.

## What It Does

During `dreego build`, this plugin runs the Tailwind CSS compiler and writes
the result to `www/static/tailwind.css`. The Go binary then embeds the CSS via
Dreego's static-asset system. No Node.js needed at runtime.

## Quick Start

In your Dreego project:

```sh
# 1. Add the plugin
go get github.com/dreego-stack/plugin-tailwind

# 2. Initialize Tailwind (one-time)
sh $(go env GOMODCACHE)/github.com/dreego-stack/plugin-tailwind@*/templates/setup.sh

# 3. Build — Tailwind CSS compiles automatically
dreego build

# 4. Include the CSS in your layout
# <link rel="stylesheet" href="/static/tailwind.css">
```

## How It Works

The plugin ships a `dreego-plugin.json` that the Dreego CLI reads during
`dreego build`. The CLI runs the configured command in your project root:

```sh
node_modules/.bin/tailwindcss -i tailwind-input.css -o www/static/tailwind.css --minify
```

This compiles Tailwind CSS, tree-shakes unused utilities, minifies, and writes
the output to `www/static/tailwind.css`. The Dreego static-asset system embeds
this file into your Go binary.

## Why node_modules/.bin (not npx)?

This plugin calls `node_modules/.bin/tailwindcss` directly instead of `npx
tailwindcss`. This is deliberate:

- **No magic**: the binary path is explicit and predictable
- **No network access**: once `npm install` has run, the binary is local
- **npm ecosystem integration**: this pattern works for any npm-packaged build
  tool, not just Tailwind — Dreego can integrate with the npm ecosystem
  without wrapping every tool in `npx`

## Files

| File | Description |
|------|-------------|
| `dreego-plugin.json` | Build-hook manifest (read by the Dreego CLI) |
| `templates/tailwind-input.css` | Tailwind entry point (copy to your project) |
| `templates/tailwind.config.js` | Default Tailwind config (copy to your project) |
| `templates/setup.sh` | One-time setup script |

## Requirements

- Node.js and npm must be installed on the build machine (not at runtime)
- `dreego build` must be the build command (not raw `go build`)

## Getting Started (Development)

```sh
make init    # download and vendor dependencies
make test    # run tests
```

## License

MPL-2.0, same as Dreego.