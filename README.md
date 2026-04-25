# searchMini Monorepo

This repository now includes a React + TypeScript frontend and a Go backend in a practical monorepo layout.

## Structure

- `/frontend` — Vite + React + TypeScript SPA
- `/packages/shared` — shared TypeScript types and interfaces
- `/src` — existing Go services and backend logic

## Local development

1. Install Node dependencies at the repo root:

```bash
npm install
```

2. Start the frontend dev server:

```bash
npm run dev
```

3. Start the backend from the Go module:

```bash
cd src/query_engine
go run .
```

4. Open the frontend at `http://localhost:4173`.

The frontend dev server proxies `/api` requests to `http://localhost:8080`.

## Build

To build both shared types and the frontend:

```bash
npm run build
```

## Deployment

### Frontend

- Build the app with `npm run build`.
- Deploy the static output from `frontend/dist` to any static host.

### Backend

- Build or deploy the Go service as normal from `src/query_engine`.
- Configure the backend to serve the production frontend build folder if you want a single host.

## Improvements included

- React + TypeScript frontend migration from the existing two HTML pages
- Tailwind CSS-based responsive UI
- Reusable components for search inputs, result cards, and theme toggling
- Shared `@searchmini/shared` package for API models
- Modern SPA routing with search results and query persistence
- Theme toggle support with localStorage

## Notes

The current backend still exposes `/api/search`, `/api/suggestions`, and `/api/random`. The frontend is designed to consume those endpoints directly.
