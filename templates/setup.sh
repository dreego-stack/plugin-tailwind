#!/bin/sh
set -e

if [ ! -f tailwind-input.css ]; then
  cp templates/tailwind-input.css tailwind-input.css
  echo "created tailwind-input.css"
fi

if [ ! -f tailwind.config.js ]; then
  cp templates/tailwind.config.js tailwind.config.js
  echo "created tailwind.config.js"
fi

mkdir -p www/static

if [ ! -f package.json ]; then
  npm init -y
fi

npm install -D tailwindcss
echo "tailwind installed to node_modules/.bin/tailwindcss"

echo ""
echo "setup complete. run 'dreego build' to compile tailwind css."