# Contributing

## Development

Requirements: Go 1.22+

```bash
go test ./...
go vet ./...
```

## UI (TypeScript)

The UI runtime assets are embedded from `internal/ui/static/`.

When you change `internal/ui/frontend/app.ts`, rebuild the generated JS:

```bash
go generate ./...
```

Or run it from the frontend folder:

```bash
cd internal/ui/frontend
npm install
npm run build
```

## i18n updates

Update `internal/i18n/locales/en.json` first, then mirror keys into `es.json` and `de.json`.
