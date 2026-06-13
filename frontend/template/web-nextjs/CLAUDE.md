# Frontend - Next.js Static Site

## Tech Stack
- **Next.js 16** (App Router, static export mode)
- **TypeScript** (strict mode)
- **Zod v4** (schema validation)
- **Axios** (HTTP client)
- **Tailwind CSS v4** (styling)

## Project Structure
```
src/
├── api/            # API layer
│   ├── client/     # Axios instance & interceptors
│   ├── request.ts  # Typed request helpers (apiGet/apiPost/...)
│   └── types/      # API-specific type definitions
├── app/            # Next.js App Router pages
├── components/     # React components
│   ├── layout/     # Layout components
│   └── ui/         # Reusable UI components
├── constants/      # App-wide constants
├── hooks/          # Custom React hooks
├── lib/            # Core utilities (validate, etc.)
├── schemas/        # Zod schemas for API validation
├── types/          # Shared TypeScript types
└── utils/          # Utility functions
```

## Commands
```bash
npm run dev           # Development server
npm run build         # Static export build → out/
npm run lint          # ESLint check
npm run lint:fix      # ESLint auto-fix
npm run format        # Prettier format
npm run format:check  # Prettier check
npm run typecheck     # TypeScript type check
```

## Architecture Notes
- **Static export**: `output: "export"` in next.config.ts — generates pure static files in `out/`
- **API validation**: All API responses are validated through Zod schemas before use
- **Backend**: Go (osbuilder/onexstack) at `NEXT_PUBLIC_API_BASE_URL` (default: http://localhost:5555)
