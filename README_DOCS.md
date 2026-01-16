# Documentation Setup

This project uses [VitePress](https://vitepress.dev/) for documentation with TypeScript and pnpm.

## Development

```bash
# Install dependencies
pnpm install

# Start dev server
pnpm run docs:dev

# Build for production
pnpm run docs:build

# Preview production build
pnpm run docs:preview
```

## GitHub Pages

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `master` branch.

The workflow is configured in `.github/workflows/docs.yml`.

## Documentation Structure

- `docs/` - Documentation source files
- `docs/.vitepress/` - VitePress configuration
- `docs/.vitepress/config.ts` - VitePress config file (TypeScript)
- `tsconfig.json` - TypeScript configuration

## Local Development

1. Install Node.js 18+ and pnpm
   ```bash
   # Install pnpm globally if not already installed
   npm install -g pnpm
   ```
2. Run `pnpm install` to install dependencies
3. Run `pnpm run docs:dev` to start the dev server
4. Open http://localhost:5173 in your browser

## Tech Stack

- **VitePress** - Static site generator
- **TypeScript** - Type-safe configuration
- **pnpm** - Fast, disk space efficient package manager
