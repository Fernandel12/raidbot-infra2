# Claude Code Helper

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
make backup-raidbot-api    # Backup main database
make backup-all            # Backup all databases
```

## Module/Package Naming
- Go module: `raidbot.app`
- API package: `rbapi`
- DB package: `rbdb`
- Domains: `raidbot.app`, `api.raidbot.app`, `community.raidbot.app`

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
go build -o bin/raidbot ./go/cmd/raidbot
```

### Run Tests
```bash
make test
# or
go test ./go/...
```

### Run the API Server
```bash
./bin/raidbot api --db-urn="..." --redis-url="localhost:6379"
```

### Use Admin Commands
```bash
./bin/raidbot admin --help
```
