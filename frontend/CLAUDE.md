# Frontend - Full Stack DevKit

## Architecture

```
frontend/
├── shared/                  # Shared code (used by both web & mobile)
│   └── src/
│       ├── api/             # Axios client factory + typed request helpers
│       ├── schemas/         # Zod schemas for API validation
│       ├── constants/       # App-wide constants
│       └── utils/           # Validation utilities
├── template/
│   ├── web-nextjs/          # Next.js 16 static site
│   └── react-native/        # React Native 0.86 mobile app
```

## Shared Package (`@devkit/shared`)

All business logic shared between web and mobile lives here:
- **API layer**: `createApiClient()`, `apiGet/Post/Put/Patch/Delete()`
- **Zod schemas**: API response validation schemas
- **Constants**: App name, API routes
- **Utils**: `validate()`, `safeValidate()`

Each app imports `@devkit/shared` and provides its own:
- Platform-specific Axios client (auth token injection)
- Platform-specific UI components
- Platform-specific routing/navigation

## Tech Stack

| Layer | Web (Next.js) | Mobile (RN) |
|-------|:---:|:---:|
| Framework | Next.js 16 | React Native 0.86 |
| Language | TypeScript | TypeScript |
| Validation | Zod v4 | Zod v4 |
| HTTP Client | Axios | Axios |
| Styling | Tailwind CSS v4 | StyleSheet |
| Build | webpack (static export) | Metro |
| Backend | Go (osbuilder/onexstack) | Go (osbuilder/onexstack) |

## Commands

### Next.js (web)
```bash
cd template/web-nextjs
npm run dev           # Dev server (with Turbopack)
npm run build         # Production build (webpack mode) → out/
npm run lint          # ESLint
npm run typecheck     # TypeScript check
```

### React Native (mobile)
```bash
cd template/react-native
npm start             # Metro bundler
npm run android       # Run on Android
npm run ios           # Run on iOS
npm run lint          # ESLint
npm run test          # Jest tests
```

## Backend

Go API server at `localhost:5555` (osbuilder/onexstack stack).
- Health check: `GET /healthz` → `{ "timestamp": "..." }`
