# RaidBot Infrastructure v2

Modern infrastructure for RaidBot

## Architecture

- **Backend**: Go with gRPC/REST API, Stripe/PayPal payments, license key management
- **Frontend**: Remix.js (TypeScript) in `/ts` directory
- **Forum**: Discourse community platform
- **Database**: MySQL + Redis
- **Deployment**: Docker with separate docker-compose per service

## Key Features

- ✅ Modern Go backend with gRPC and REST API support
- ✅ PayPal payment integration
- ✅ License key generation and management
- ✅ Discourse SSO integration
- ✅ Discord bot integration for role management
- ✅ Admin tools for user and license management

## Project Structure

```
raidbot-infra2/
├── api/
│   ├── proto/raidbot/          # Protobuf definitions
│   │   ├── rbdb.proto          # Database models
│   │   ├── rbapi.proto         # API service definitions
│   │   └── errcode.proto       # Error codes
│   ├── buf.yaml                # Buf configuration
│   └── buf.gen.yaml            # Buf generation config
├── go/
│   ├── cmd/raidbot/            # CLI entrypoint
│   │   ├── main.go
│   │   ├── api.go              # API server command
│   │   ├── admin.go            # Admin commands
│   │   ├── cli.go              # Test client
│   │   └── token.go            # SSO token handling
│   ├── pkg/rbapi/              # API implementation
│   │   ├── server.go           # HTTP/gRPC server
│   │   ├── service.go          # Business logic
│   │   ├── admin_*.go          # Admin endpoints
│   │   ├── payment_*.go        # Payment endpoints
│   │   ├── license_*.go        # License endpoints
│   │   ├── user_*.go           # User endpoints
│   │   └── ...
│   ├── pkg/rbdb/               # Generated DB models
│   ├── pkg/errcode/            # Generated error codes
│   └── internal/jsonutil/      # JSON utilities
├── ts/                         # Frontend (Remix.js + TypeScript)
│   ├── app/                    # Remix application
│   │   ├── routes/             # Page routes
│   │   ├── components/         # React components
│   │   ├── i18n/               # Internationalization
│   │   └── lib/                # Utilities and API clients
│   ├── functions/              # Cloudflare Pages functions
│   ├── public/                 # Static assets
│   └── package.json            # Node dependencies
├── deployments/
│   ├── raidbot-api/            # API service deployment
│   ├── nginx-proxy/            # Nginx reverse proxy
│   └── discourse.yml           # Discourse configuration
├── go.mod                      # Go module file
├── Makefile                    # Build commands
├── Dockerfile                  # Multi-stage Docker build
└── README.md                   # This file
```

## Quick Start

### Prerequisites

- Go 1.23.6+
- Node.js 18+ and pnpm
- Docker and Docker Compose
- Buf CLI (for protobuf generation)
- Make

### Development Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd raidbot-infra2
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Generate protobuf code**:
   ```bash
   make buf-generate
   ```

4. **Build the binary**:
   ```bash
   make build
   ```

5. **Run tests**:
   ```bash
   make test
   ```

6. **Setup frontend** (optional):
   ```bash
   cd ts
   pnpm install
   pnpm dev
   ```

## Development Commands

### Protobuf Generation

After making changes to `.proto` files:

```bash
make buf-generate
```

This will:
1. Generate Go code from protobuf definitions
2. Move generated files to appropriate packages
3. Apply necessary transformations

### Building

```bash
# Build the binary
make build

# Build Docker image
make docker.build

# Build and push Docker image
make docker.push
```

### Testing

```bash
# Run backend tests
make test

# Run with coverage
make test-coverage

# Run frontend typecheck
cd ts && pnpm typecheck

# Run frontend linter
cd ts && pnpm lint
```

### Database Operations

```bash
# Backup main database
make backup-raidbot-api

# Backup all databases
make backup-all
```

## Deployment

### Production Deployment

1. **Set up networks**:
   ```bash
   docker network create service-proxy
   ```

2. **Deploy nginx-proxy** (first time only):
   ```bash
   cd deployments/nginx-proxy
   docker compose up -d
   ```

3. **Deploy Discourse** (first time only):
   ```bash
   cd deployments
   # Edit discourse.yml with your settings
   ./launcher bootstrap discourse
   ./launcher start discourse
   ```

4. **Deploy RaidBot API**:
   ```bash
   cd deployments/raidbot-api
   cp .env.example .env
   # Edit .env with your credentials
   docker compose up -d
   ```

### Environment Variables

Create a `.env` file in `deployments/raidbot-api/`:

```env
# Database
URN=raidbot:password@tcp(mysql:3306)/raidbot?charset=utf8&parseTime=True&loc=Local
MYSQL_PASSWORD=your_mysql_password

# PayPal
PAYPAL_CLIENT_ID=...
PAYPAL_CLIENT_SECRET=...
PAYPAL_WEBHOOK_ID=...

# Discord
DISCORD_BOT_TOKEN=...
```

## CLI Usage

### Admin Commands

```bash
# Get active users
./raidbot admin active-users --server localhost:8080

# Create license for user
./raidbot admin create-license \
  --user-email user@example.com \
  --duration ONE_MONTH

# Revoke license
./raidbot admin revoke-license --key LICENSE_KEY_HERE

# Search database
./raidbot admin search --term "search_term"
```

### Testing User Session

```bash
# Test HTTP endpoint
./raidbot cli @me --server http://localhost:8080

# Test gRPC endpoint
./raidbot cli @me --grpc --server localhost:8080
```

## Module and Package Naming

- Go module: `raidbot.app`
- API package: `rbapi`
- DB package: `rbdb`
- Domains:
  - Main site: `raidbot.app`
  - API: `api.raidbot.app`
  - Community: `community.raidbot.app`

## Database Schema

The database schema is defined in `api/proto/raidbot/rbdb.proto` and includes:

- **User** - User accounts (synced with Discourse)
- **LicenseKey** - Software license keys
- **Payment** - Payment records (Stripe/PayPal)
- **Subscription** - Recurring subscriptions
- **Activity** - Audit log for all operations

## SSO Integration

RaidBot uses Discourse for authentication via SSO:

1. User clicks login on frontend
2. Frontend redirects to API SSO endpoint
3. API redirects to Discourse with signed payload
4. User authenticates on Discourse
5. Discourse redirects back with user info
6. API creates session and redirects to frontend

## Payment Flow

### One-time Purchase

1. User selects license duration
2. Frontend calls `/payment/paypal/create-checkout`
3. User completes payment on Stripe/PayPal
4. Webhook receives payment confirmation
5. License key is generated and assigned to user
6. User receives confirmation email

## License Management

### Admin License Creation

Admins can create licenses manually:

```bash
./raidbot admin create-license \
  --user-email user@example.com \
  --duration LIFETIME
```

## Monitoring and Maintenance

### View Logs

```bash
cd deployments/raidbot-api
docker compose logs -f
```

### Backup Database

```bash
cd deployments/raidbot-api
./db_backup.sh
```

Backups are stored in `deployments/raidbot-api/backups/` and automatically compressed. Old backups (30+ days) are automatically deleted.

### Restart Services

```bash
cd deployments/raidbot-api
docker compose restart
```

## Troubleshooting

### Protobuf Generation Issues

If `make buf-generate` fails:

```bash
# Install buf CLI
go install github.com/bufbuild/buf/cmd/buf@latest

# Install protoc-gen-gorm
go install github.com/infobloxopen/protoc-gen-gorm@latest

# Try again
make buf-generate
```

### Database Connection Issues

Check the URN format in `.env`:

```
URN=username:password@tcp(hostname:port)/database?charset=utf8&parseTime=True&loc=Local
```

### SSL Certificate Issues

The nginx-proxy automatically requests Let's Encrypt certificates. Ensure:
- DNS records point to your server
- Ports 80 and 443 are open
- `LETSENCRYPT_HOST` is set correctly in docker-compose.yml

## Contributing

1. Create a feature branch
2. Make changes
3. Run tests: `make test`
4. Update protobuf if needed: `make buf-generate`
5. Submit pull request

## License

Proprietary - All rights reserved

## Support

For issues and questions:
- Discord: [Join our community](https://discord.gg/raidbot)
- Forum: https://community.raidbot.app
- Email: support@raidbot.app
