# Agent Instructions for plugin-tailwind

- Don't create binaries here — only in /tmp or ./tmp

## Language Rule

- **Chat with user:** German
- **Everything in this repository:** English (code, comments, docs, commits, tests)

## What This Is

This is the **Tailwind CSS plugin** for the Dreego framework. It is a
**build-time** plugin — it integrates Tailwind CSS compilation into the
`dreego build` pipeline via the Dreego build-hook mechanism. It does NOT
register any runtime routes or middleware.

## How It Works

1. A project adds this plugin: `go get github.com/dreego-stack/plugin-tailwind`
2. The project runs `sh templates/setup.sh` once to create `tailwind-input.css`,
   `tailwind.config.js`, and install Tailwind via npm
3. During `dreego build`, the CLI reads `dreego-plugin.json` and runs
   `node_modules/.bin/tailwindcss -i tailwind-input.css -o www/static/tailwind.css --minify`
4. The compiled CSS lands in `www/static/tailwind.css`
5. Dreego's static-asset system embeds the CSS into the Go binary
6. The app includes the CSS via `<link rel="stylesheet" href="/static/tailwind.css">`

## npm Integration

This plugin uses `node_modules/.bin/tailwindcss` directly (not `npx`). The
setup.sh installs Tailwind via `npm install -D tailwindcss`, which places the
binary in `node_modules/.bin/`. The build-hook calls that path directly. This
approach lets Dreego interact with the npm ecosystem without magic wrappers.

## Build-Hook Contract

The `dreego-plugin.json` at the repo root defines the build steps:

```json
{
  "build": {
    "steps": [
      {
        "cmd": "node_modules/.bin/tailwindcss -i tailwind-input.css -o www/static/tailwind.css --minify",
        "when": "pre-build"
      }
    ]
  }
}
```

The `cmd` runs in the **project root** (where go.mod lives), not in the plugin
directory. See the Dreego build-hooks documentation:
https://github.com/dreego-stack/dreego/blob/main/_docs/build-hooks.md

## Testing

- `go test ./...` — unit tests (minimal; this is a build-time plugin)
- The build-hook mechanism is tested in the Dreego main repo's CLI tests

## CI

- `.github/workflows/ci.yml` — `go vet`, `go test -race`, and a compatibility
  job that tests against the latest published dreego tag
- `.github/workflows/release.yml` — validates change file, tests, creates tag
- `.github/dependabot.yml` — auto-updates dreego dependency weekly

## Coding Rules

- Max 300 lines per handwritten file
- No code comments (except where needed for clarity)
- Go 1.27+, prefer standard library
- One Go package per repository

## Commit Convention

Every change lands via a pull request with one `.changes/*.md` file:

```yaml
---
version: patch
---

- Feat: add X
- Bug: fix Y
```
