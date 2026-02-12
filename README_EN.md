<p align="center">
  <img src="https://img.shields.io/badge/A4AD-Profile%20Service-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="A4AD Profile Service">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/PostgreSQL-16+-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/Status-Early%20Development-orange?style=flat-square" alt="Status">
</p>

---

## 📝 Overview

**Profile Service** manages user profiles, public information, and statistics for the A4AD Forum. It provides APIs for viewing and editing profiles, tracking user activity metrics (posts, comments, reputation), and handling avatar uploads.

---

## ✨ Features

- 👤 **User Profiles** — Public profile pages with customizable information
- 🖼️ **Avatar Management** — Upload and manage profile pictures
- 📝 **Bio & Location** — Personal description and location tags
- 📊 **Activity Stats** — Track posts count, comments count, and reputation score
- 🏅 **Reputation System** — Community-driven reputation points
- 🔍 **Profile Search** — Find users by username or bio keywords
- 🛡️ **Privacy Controls** — Configurable profile visibility settings
- 📅 **Join Date Tracking** — Member since timestamp display

---

## 🛠 Tech Stack

- **Language:** Go 1.23+
- **Framework:** Standard library + Chi router (or Fiber)
- **Database:** PostgreSQL 16+
- **ORM/Query Builder:** sqlc or GORM
- **Migrations:** golang-migrate
- **File Storage:** Local filesystem or S3-compatible (MinIO)
- **Image Processing:** Imaging or bimg library
- **Validation:** go-playground/validator
- **Configuration:** Viper

---

## 📊 Architecture

```mermaid
flowchart LR
    Gateway["API Gateway<br/>Go + Fiber"] --> Profile["Profile Service<br/>Go"]
    Profile --> PostgreSQL[("PostgreSQL<br/>Profiles Table")]
    Profile --> Storage[("Object Storage<br/>Avatars")]
    
    Auth["Auth Service<br/>Java"] -.->|"User events"| Profile
    Post["Post Service<br/>Go"] -.->|"Post count updates"| Profile
    Comment["Comment Service<br/>NestJS"] -.->|"Comment count updates"| Profile
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/A4AD-team/profile-service.git
   cd profile-service
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Start PostgreSQL:**
   ```bash
   docker compose up -d postgres
   ```

4. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. **Run database migrations:**
   ```bash
   make migrate-up
   # Or: migrate -path ./migrations -database "postgres://user:pass@localhost:5432/profile_db?sslmode=disable" up
   ```

6. **Run the service:**
   ```bash
   # Development mode
   go run ./cmd/server

   # Or with custom config
   go run ./cmd/server -config=./config/local.yaml
   ```

The service will be available at `http://localhost:8082`

---

## 🔧 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_ENV` | Environment (development, staging, production) | `development` | No |
| `APP_PORT` | HTTP server port | `8082` | No |
| `APP_HOST` | HTTP server host | `0.0.0.0` | No |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/profile_db?sslmode=disable` | Yes |
| `DATABASE_MAX_CONNS` | Max database connections | `25` | No |
| `DATABASE_MIN_CONNS` | Min database connections | `5` | No |
| `STORAGE_TYPE` | Storage type (local, s3) | `local` | No |
| `STORAGE_LOCAL_PATH` | Local storage path | `./uploads` | No |
| `S3_ENDPOINT` | S3-compatible endpoint | `` | If using S3 |
| `S3_BUCKET` | S3 bucket name | `` | If using S3 |
| `S3_ACCESS_KEY` | S3 access key | `` | If using S3 |
| `S3_SECRET_KEY` | S3 secret key | `` | If using S3 |
| `MAX_AVATAR_SIZE` | Max avatar file size in bytes | `2097152` (2MB) | No |
| `ALLOWED_AVATAR_TYPES` | Allowed MIME types | `image/jpeg,image/png,image/webp` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `METRICS_ENABLED` | Enable Prometheus metrics | `true` | No |

---

## 📡 API Endpoints

### Public Routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/profiles/:username` | Get public profile by username |
| `GET` | `/api/v1/profiles/:username/avatar` | Get user's avatar |
| `GET` | `/api/v1/profiles/search` | Search profiles (query: `?q=keyword`) |

### Protected Routes (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/profiles/me` | Get current user's profile |
| `PATCH` | `/api/v1/profiles/me` | Update current user's profile |
| `POST` | `/api/v1/profiles/me/avatar` | Upload new avatar |
| `DELETE` | `/api/v1/profiles/me/avatar` | Remove avatar (reset to default) |
| `GET` | `/api/v1/profiles/me/stats` | Get detailed activity statistics |

### Admin Routes

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/admin/profiles` | List all profiles (with filters) |
| `PATCH` | `/api/v1/admin/profiles/:id/reputation` | Adjust user reputation |

### Example Profile Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "user-uuid",
  "username": "johndoe",
  "avatarUrl": "https://cdn.example.com/avatars/johndoe.png",
  "bio": "Full-stack developer passionate about Go and open source",
  "location": "San Francisco, CA",
  "joinedAt": "2026-01-15T10:30:00Z",
  "stats": {
    "postCount": 42,
    "commentCount": 156,
    "reputation": 1280
  },
  "isPublic": true
}
```

---

## 🩺 Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | `GET` | Overall health status |
| `/health/live` | `GET` | Liveness probe |
| `/health/ready` | `GET` | Readiness probe (checks database) |
| `/metrics` | `GET` | Prometheus metrics |

### Example Response

```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T15:30:00Z",
  "version": "0.1.0",
  "checks": {
    "database": "healthy",
    "storage": "healthy"
  }
}
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestProfileService ./...

# Run integration tests
go test -tags=integration ./...
```

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>Built with ❤️ by A4AD Team</strong>
</p>
