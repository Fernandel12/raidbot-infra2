# Claude Code Helper

## Git Commit Guidelines

- Do NOT use conventional commits (no `feat:`, `fix:`, `chore:` prefixes)
- Do NOT add "Generated with Claude Code" footer
- Do NOT add "Co-Authored-By: Claude" footer

## Project Overview

This is raidbot-infra2 - infrastructure for RaidBot, a bot for RAID: Shadow Legends game.

### Architecture
- **Backend**: Go with gRPC/REST API, Stripe/PayPal payments, license key management ✅
- **Frontend**: Remix.js (integrated in `/ts` directory) ✅
- **Forum**: Discourse community platform
- **Database**: MySQL + Redis
- **Deployment**: Docker with separate docker-compose per service

### Key Features
- Modern Go backend with payment integration
- Discourse SSO integration for authentication
- License key generation and management
- Discord bot integration for role management

## Project Commands

### Generate Protobuf Code
After making changes to any `.proto` files in the `api/proto` directory, run:
```bash
cd raidbot-infra2
make buf-generate
```

This command will:
1. Generate Go code from the protobuf definitions
2. Move generated files to appropriate packages
3. Apply necessary transformations

### Run Tests
```bash
make test
```

### Build Docker Image
```bash
make docker.build
```

### Database Backups
```bash
make backup-rslbot-api    # Backup main database
make backup-all            # Backup all databases
```

## Module/Package Naming
- Go module: `rslbot.com`
- API package: `rbapi`
- DB package: `rbdb`
- Domains: `rslbot.com`, `api.rslbot.com`, `community.rslbot.com`

## Implementation Progress

### Completed ✓
1. Go backend with gRPC/REST API, payment integration, license management
2. Protobuf API definitions (rbdb.proto, rbapi.proto, errcode.proto)
3. Remix.js frontend with TypeScript
4. Discourse SSO integration for authentication
5. Stripe and PayPal payment integration
6. Docker deployment configurations
7. MySQL + Redis database setup

### Status: ✅ OPERATIONAL

**Backend:**
- Protobuf code generation (`make buf-generate`)
- Building (`make build`)
- Testing (`make test`)
- Running API server

**Frontend:**
- Located in `/ts` directory
- Development server (`cd ts && pnpm dev`)
- Type checking (`cd ts && pnpm typecheck`)
- Production build (`cd ts && pnpm build`)

## API Structure

Core services:
- **Admin APIs**: AddLicenseKey, RevokeLicense, SearchDatabase, GetActiveUsers
- **Payment APIs**: CreateStripeCheckout, CreatePayPalCheckout (+ webhooks)
- **User APIs**: GetSession, GetLicenses, Logout, SyncDiscordRole
- **Public APIs**: ToolStatus

Database entities:
- User (with discourse_id for SSO)
- LicenseKey (key, duration, revoked, user_id)
- Payment (provider, reference_id, amount, license_duration, user_id)
- Activity (audit log for all operations)

## Build Instructions

The project is now fully set up and builds successfully. To work with it:

### Generate Protobuf Code
```bash
make buf-generate
```

### Build Binary
```bash
make build
# or
go build -o bin/rslbot ./go/cmd/rslbot
```

### Run Tests
```bash
make test
# or
go test ./go/...
```

### Use Admin Commands
```bash
./bin/rslbot admin --help
```

## Local Development

### Start Backend (API + MySQL + Redis)
```bash
cd go
export PATH="$PATH:$(go env GOPATH)/bin"
make api
```

This will:
1. Start MySQL and Redis containers using `docker-compose.yml` + `docker-compose.dev.yml` in `/go`
2. Install CompileDaemon for hot-reload
3. Start the API server on `:8080` with auto-rebuild on file changes

**IMPORTANT**: Use the docker-compose files in `/go` directory for development. Do NOT use the production docker-compose in `/deployments/rslbot-api/`.

### Start Frontend
```bash
cd ts
pnpm dev
```

This starts the Remix dev server on `http://localhost:5173/` (or next available port).

### Verify Database
```bash
cd go
docker-compose -p rslbot -f docker-compose.yml -f docker-compose.dev.yml exec -T mysql mysql -u root -puns3cur3 rslbot -e "SHOW TABLES;"
```

Expected tables: `activities`, `license_keys`, `payments`, `subscriptions`, `users`
