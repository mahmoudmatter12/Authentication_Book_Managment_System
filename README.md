```markdown
# Go Authentication & Book Management System

![Go](https://img.shields.io/badge/Go-1.21+-blue)
![Gin](https://img.shields.io/badge/Gin-v1.9-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blueviolet)

A production-ready authentication system with complete CRUD operations for Users and Books, featuring:

- JWT Authentication
- Role-based access control (User/Admin)
- Structured logging with Zap
- Rate limiting
- Graceful shutdown
- Health checks
- CORS support

## 📦 Dependencies

### Core Packages
| Package | Description | Installation |
|---------|-------------|--------------|
| Gin | HTTP Web Framework | `go get -u github.com/gin-gonic/gin` |
| GORM | ORM for Go | `go get -u gorm.io/gorm` |
| PostgreSQL Driver | GORM PostgreSQL adapter | `go get -u gorm.io/driver/postgresql` |
| JWT | JSON Web Tokens | `go get -u github.com/golang-jwt/jwt/v5` |
| Zap | Structured Logging | `go get -u go.uber.org/zap` |
| godotenv | Environment Variables | `go get -u github.com/joho/godotenv` |
| cors | CORS Middleware | `go get -u github.com/gin-contrib/cors` |
| requestid | Request Tracing | `go get -u github.com/gin-contrib/requestid` |

### Development Tools
| Tool | Purpose | Installation |
|------|---------|--------------|
| CompileDaemon | Auto-reload during development | `go install github.com/githubnemo/CompileDaemon@latest` |
| Migrate | Database migrations (optional) | `brew install golang-migrate` (Mac) or see [install docs](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) |

## 🚀 Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/auth-system.git
   cd auth-system
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL**
   ```bash
   # Mac (Homebrew)
   brew install postgresql
   brew services start postgresql

   # Ubuntu
   sudo apt update && sudo apt install postgresql postgresql-contrib
   sudo service postgresql start
   ```

4. **Create database**
   ```bash
   psql -U postgres -c "CREATE DATABASE authsystem;"
   ```

5. **Configure environment**
   Create `.env` file:
   ```env
   # Database
   DB_DSN="host=localhost user=postgres password=yourpassword dbname=authsystem port=5432 sslmode=disable"

   # JWT
   JWT_SECRET=your_secure_secret_key_here

   # Server
   PORT=8080
   GIN_MODE=debug
   FRONTEND_URL=http://localhost:3000
   APP_VERSION=1.0.0
   ```

## 🏃 Running the Application

### Development Mode (with auto-reload)
```bash
CompileDaemon --command="./authSystem"
```

### Production Mode
```bash
go build -o authSystem && ./authSystem
```

## 🌐 API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/signup` | Register new user |
| POST | `/auth/login` | Login with credentials |
| GET | `/auth/validate` | Validate JWT token |
| GET | `/auth/logout` | Invalidate JWT token |

### Books (Require Authentication)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/book/:id` | Get single book |
| POST | `/api/book` | Create new book |
| PATCH | `/api/book/:id` | Update book |
| DELETE | `/api/book/:id` | Delete book |

### Admin Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/users` | List all users (Admin only) |
| GET | `/admin/books` | List all books (Admin only) |

## 📊 Example Requests

### User Registration
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123"}'
```

### Book Creation
```bash
curl -X POST http://localhost:8080/api/book \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Go Programming","author":"John Doe"}'
```

## 🛠 Project Structure

```
authSystem/
├── controllers/      # Business logic handlers
│   ├── auth.go       # Authentication controllers
│   ├── book.go       # Book management
│   └── user.go       # User management
├── initializers/     # Startup/config
│   ├── database.go   # DB connection
│   └── env.go        # Environment config
├── middleware/       # HTTP middleware
│   ├── auth.go       # JWT authentication
│   ├── logger.go     # Request logging
│   └── rate_limit.go # Rate limiting
├── models/           # Database models
│   ├── book.go       # Book model
│   └── user.go       # User model
├── go.mod           # Go module file
├── go.sum           # Dependency checksums
└── main.go          # Application entrypoint
```
