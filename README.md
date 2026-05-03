# VogueLuxe — E-Commerce Platform (Microservices Architecture)

> **🔀 Architecture:** This is the **microservices** version of VogueLuxe, decomposed from the [monolithic codebase](https://github.com/Mohammad-Niyas/VogueLuxe-Ecommerce-Monolithic) using the **Strangler Fig pattern**. Domain services (Wallet, Wishlist) have been extracted into independent gRPC microservices, while the core e-commerce logic remains in a modular monolith that acts as both the primary application and the gRPC client.

---

## Architecture Overview

```
                         ┌─────────────────────────┐
                         │      API Gateway         │
                         │      (Gin - :3000)       │
                         │   Reverse Proxy to       │
                         │   Monolith Core          │
                         └────────────┬─────────────┘
                                      │ HTTP
                                      ▼
              ┌───────────────────────────────────────────────────┐
              │              MONOLITH CORE (:8080)                │
              │                                                   │
              │   ┌──────────┐ ┌──────────┐ ┌────────────────┐   │
              │   │  Auth &  │ │ Catalog  │ │    Orders &    │   │
              │   │  Users   │ │ Products │ │    Payments    │   │
              │   │  (JWT,   │ │ Category │ │   (Razorpay,   │   │
              │   │  OAuth,  │ │ Variants │ │    Invoice)    │   │
              │   │  OTP)    │ │ Offers   │ │                │   │
              │   └──────────┘ └──────────┘ └────────────────┘   │
              │                                                   │
              │   ┌──────────────────┐ ┌──────────────────────┐  │
              │   │   gRPC Client:   │ │   gRPC Client:       │  │
              │   │   Wishlist       │ │   Wallet             │  │
              │   │   (:50051)       │ │   (:8082)            │  │
              │   └────────┬─────────┘ └────────┬─────────────┘  │
              │            │                    │                 │
              └────────────┼────────────────────┼─────────────────┘
                           │ gRPC               │ gRPC
                           ▼                    ▼
              ┌────────────────────┐  ┌────────────────────────┐
              │  WISHLIST SERVICE  │  │    WALLET SERVICE       │
              │     (:50051)       │  │       (:8082)           │
              │                    │  │                         │
              │  • Own PostgreSQL  │  │  • Own PostgreSQL       │
              │    (wishlist DB)   │  │    (wallet_db)          │
              │  • Own Models      │  │  • Own Models           │
              │  • gRPC Server     │  │  • gRPC Server          │
              │  • JWT Interceptor │  │  • JWT Interceptor      │
              │  • Logging         │  │  • Logging              │
              │    Interceptor     │  │    Interceptor          │
              └────────────────────┘  │  • Razorpay Integration │
                                      └────────────────────────┘
```

### Decomposition Strategy: Strangler Fig Pattern

The migration from monolith to microservices follows the **Strangler Fig** approach — instead of rewriting everything at once, individual bounded contexts are extracted one at a time:

| Phase | Status | What Changed |
| :--- | :---: | :--- |
| **Phase 0** | ✅ Done | Original monolith — all domains in one binary |
| **Phase 1** | ✅ Done | Wishlist Service extracted → gRPC server on `:50051`, own database |
| **Phase 2** | ✅ Done | Wallet Service extracted → gRPC server on `:8082`, own database with Razorpay |
| **Phase 3** | 🔜 Next | API Gateway introduced → Reverse proxy routing traffic to monolith |
| **Phase 4** | 📋 Planned | Auth, Catalog, Order services to be extracted next |

---

## Tech Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go 1.25 | All services written in Go |
| **Web Framework** | Gin-Gonic | HTTP routing in Monolith Core & API Gateway |
| **Inter-Service Comm** | gRPC + Protocol Buffers | Strongly-typed, high-performance RPC between services |
| **ORM** | GORM | PostgreSQL interactions with AutoMigrate |
| **Database** | PostgreSQL (per-service) | Each microservice owns its database (Database-per-Service pattern) |
| **Auth** | JWT (HS256) + Google OAuth 2.0 | Shared JWT secret across services; token passed via gRPC metadata |
| **Payments** | Razorpay API | Wallet top-up & checkout payments |
| **Media** | Cloudinary | Product image CDN & storage |
| **Email/OTP** | Brevo (Sendinblue) | Transactional emails and OTP |
| **PDF Generation** | gofpdf | Invoice & sales report export |
| **Frontend** | Go `html/template` + Tailwind CSS | Server-side rendered templates |

---

## Project Structure

```
vogueluxe-microservices/
│
├── api-gateway/                          # ENTRY POINT — Reverse Proxy
│   ├── main.go                           # Gin server on :3000, proxies all traffic to :8080
│   └── go.mod
│
├── monolithic/                           # CORE SERVICE — The modular monolith
│   ├── main.go                           # Starts Gin on :8080, initializes gRPC clients
│   ├── config/
│   │   ├── dbConnection.go               # PostgreSQL connection (main e-commerce DB)
│   │   ├── envLoad.go                    # godotenv loader
│   │   └── googleAuth.go                # Google OAuth2 config
│   ├── controllers/
│   │   ├── AdminController.go            # Dashboard, reports, PDF/Excel exports
│   │   ├── AdminOrderController.go       # Order status, return approvals
│   │   ├── CategoryController.go         # Category CRUD, category offers
│   │   ├── CouponController.go           # Coupon CRUD, validation
│   │   ├── ProductController.go          # Product/variant CRUD, Cloudinary uploads
│   │   ├── UserAddtoCart.go              # Cart, checkout, order creation
│   │   ├── UserController.go             # Auth, profile, addresses, OAuth
│   │   └── WalletController.go           # Delegates to Wallet Service via gRPC
│   ├── middleware/
│   │   └── jwt.go                        # JWT auth, RBAC, no-cache headers
│   ├── models/                           # GORM entities for core domain
│   ├── routers/
│   │   ├── Admin_Router.go               # ~50 admin routes
│   │   └── User_Router.go                # ~50 user routes (includes wallet & wishlist proxies)
│   ├── pkg/
│   │   ├── client/
│   │   │   ├── wishlist.go               # gRPC client → Wishlist Service (:50051)
│   │   │   └── wallet.go                 # gRPC client → Wallet Service (:8082)
│   │   └── pb/
│   │       └── wallet/                   # Generated protobuf Go code for Wallet
│   ├── utils/
│   │   ├── Cloudinary.go                 # Image upload/delete
│   │   ├── Otp.go                        # OTP via Brevo API
│   │   └── paymentRazorpay.go            # Razorpay order creation
│   └── views/
│       ├── Admin/                        # 17 admin SSR templates
│       └── User/                         # 23 user SSR templates
│
├── wishlist-service/                     # MICROSERVICE — Wishlist Domain
│   ├── main.go                           # gRPC server on :50051
│   ├── proto/
│   │   └── wishlist.proto                # Service contract (AddToWishlist, Remove, Get)
│   ├── controllers/
│   │   └── wishlist.go                   # gRPC handler implementations
│   ├── models/
│   │   └── wishlist.go                   # Wishlist, WishlistItem (own DB tables)
│   ├── middleware/
│   │   ├── auth.go                       # JWT gRPC unary interceptor
│   │   └── logging.go                    # Request logging interceptor
│   ├── config/
│   │   └── db.go                         # PostgreSQL connection (wishlist DB)
│   ├── pkg/pb/                           # Generated protobuf Go code
│   └── views/
│       └── User_Profile_Wishlist.html    # SSR template for wishlist page
│
└── wallet-service/                       # MICROSERVICE — Wallet Domain
    ├── main.go                           # gRPC server on :8082
    ├── proto/
    │   └── wallet.proto                  # Service contract (GetWallet, CreateTransaction,
    │                                     #   ConfirmPayment, DebitWallet)
    ├── controllers/
    │   └── wallet.go                     # gRPC handler implementations
    ├── models/
    │   └── wallet.go                     # Wallet, WalletTransaction (own DB tables)
    ├── middleware/
    │   ├── auth.go                       # JWT gRPC unary interceptor
    │   └── logging.go                    # Request logging interceptor
    ├── config/
    │   └── database.go                   # PostgreSQL connection (wallet_db)
    └── pkg/pb/                           # Generated protobuf Go code
```

---

## Service Contracts (gRPC / Protocol Buffers)

### Wishlist Service — `wishlist.proto`

| RPC Method | Request | Response | Description |
| :--- | :--- | :--- | :--- |
| `AddToWishlist` | `user_id`, `product_id`, `variant_id` | `status`, `message` | Add product variant to user's wishlist |
| `RemoveFromWishlist` | `user_id`, `wishlist_item_id` | `status`, `message` | Remove item from wishlist |
| `GetWishlist` | `user_id` | `repeated WishlistItem` | Fetch all wishlist items for a user |

### Wallet Service — `wallet.proto`

| RPC Method | Request | Response | Description |
| :--- | :--- | :--- | :--- |
| `GetWallet` | `user_id` | `Wallet`, `repeated Transaction` | Get wallet balance and transaction history |
| `CreateTransaction` | `user_id`, `amount` | Razorpay order details + `transaction_id` | Initiate wallet top-up via Razorpay |
| `ConfirmPayment` | `user_id`, `transaction_id`, Razorpay verification fields | `status`, `redirect_url` | Verify Razorpay signature, credit wallet |
| `DebitWallet` | `user_id`, `amount`, `order_id` | `status`, `transaction_id` | Deduct balance during checkout |

---

## Inter-Service Communication

```
┌──────────────────────────────────────────────────────────────┐
│                    REQUEST FLOW EXAMPLE                       │
│                  (User adds to wishlist)                      │
│                                                              │
│  Browser ──HTTP──▶ API Gateway (:3000)                       │
│                       │                                      │
│                       │ reverse proxy                        │
│                       ▼                                      │
│              Monolith Core (:8080)                            │
│                       │                                      │
│                       │ 1. JWT middleware validates token     │
│                       │ 2. Controller extracts user_id       │
│                       │ 3. Calls gRPC client                 │
│                       ▼                                      │
│              WishlistClient.AddToWishlist()                   │
│                       │                                      │
│                       │ gRPC (protobuf over HTTP/2)          │
│                       │ + JWT token in metadata              │
│                       ▼                                      │
│              Wishlist Service (:50051)                        │
│                       │                                      │
│                       │ 1. Logging interceptor               │
│                       │ 2. JWT interceptor validates token   │
│                       │ 3. Handler queries wishlist DB       │
│                       │ 4. Returns protobuf response         │
│                       ▼                                      │
│              Response flows back to browser                   │
└──────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

| Decision | Implementation |
| :--- | :--- |
| **Database-per-Service** | `monolithic` → main e-commerce DB, `wishlist-service` → `wishlist` DB, `wallet-service` → `wallet_db` |
| **gRPC over REST** | Internal service-to-service communication uses gRPC for type safety, code generation, and HTTP/2 performance |
| **Shared JWT Secret** | All services validate the same JWT tokens — the monolith issues tokens, microservices verify via gRPC interceptors |
| **gRPC Interceptor Chain** | Each microservice uses `grpc.ChainUnaryInterceptor` with Logging → JWT auth middleware |
| **Module Replace Directive** | The monolith uses `go.mod replace` to reference the wishlist service's protobuf package locally |

---

## Databases

Each service owns its data — no cross-service database access:

```
┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐
│   Main E-Commerce   │  │     wishlist DB      │  │     wallet_db       │
│     PostgreSQL      │  │     PostgreSQL       │  │     PostgreSQL      │
│                     │  │                      │  │                     │
│  • Users            │  │  • Wishlist          │  │  • Wallet           │
│  • Products         │  │  • WishlistItem      │  │  • WalletTransaction│
│  • Categories       │  │                      │  │                     │
│  • Orders           │  │  Owner:              │  │  Owner:             │
│  • Carts            │  │  wishlist-service     │  │  wallet-service     │
│  • Coupons          │  │                      │  │                     │
│  • Addresses        │  └─────────────────────┘  └─────────────────────┘
│  • Payments         │
│  • OTPs             │
│                     │
│  Owner:             │
│  monolithic (core)  │
└─────────────────────┘
```

---

## Getting Started

### Prerequisites

- **Go 1.25+**
- **PostgreSQL** — 3 databases: main e-commerce DB, `wishlist`, `wallet_db`
- **protoc** + `protoc-gen-go` + `protoc-gen-go-grpc` (for regenerating protobuf code)

### 1. Setup Databases

Create three PostgreSQL databases:

```sql
CREATE DATABASE ecommerce;
CREATE DATABASE wishlist;
CREATE DATABASE wallet_db;
```

### 2. Configure Environment Variables

**Monolith Core** (`monolithic/.env`):
```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ecommerce
DB_PORT=5432
SECRETKEY=your_jwt_secret

GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
BREVO_API_KEY=...
RAZORPAY_KEY_ID=...
RAZORPAY_KEY_SECRET=...
```

**Wishlist Service** (`wishlist-service/.env`):
```env
SECRETKEY=your_jwt_secret
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=wishlist
DB_PORT=5432
```

**Wallet Service** (`wallet-service/.env`):
```env
PORT=8082
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=wallet_db
DB_PORT=5432
RAZORPAY_KEY_ID=...
RAZORPAY_KEY_SECRET=...
SECRETKEY=your_jwt_secret
```

### 3. Start All Services

Run each service in a separate terminal:

```bash
# Terminal 1: Wishlist Service (gRPC :50051)
cd wishlist-service
go mod tidy && go run main.go

# Terminal 2: Wallet Service (gRPC :8082)
cd wallet-service
go mod tidy && go run main.go

# Terminal 3: Monolith Core (HTTP :8080)
cd monolithic
go mod tidy && go run main.go

# Terminal 4: API Gateway (HTTP :3000)
cd api-gateway
go mod tidy && go run main.go
```

### 4. Access the Application

- **Via Gateway**: http://localhost:3000
- **Direct to Monolith**: http://localhost:8080

---

## Regenerating Protobuf Code

If you modify `.proto` files, regenerate the Go code:

```bash
# Wishlist
cd wishlist-service
protoc --go_out=. --go-grpc_out=. proto/wishlist.proto

# Wallet
cd wallet-service
protoc --go_out=. --go-grpc_out=. proto/wallet.proto
```

---

## Service Ports

| Service | Protocol | Port | Description |
| :--- | :--- | :--- | :--- |
| API Gateway | HTTP | `:3000` | Entry point, reverse proxy |
| Monolith Core | HTTP | `:8080` | Main application (SSR + REST) |
| Wishlist Service | gRPC | `:50051` | Wishlist domain operations |
| Wallet Service | gRPC | `:8082` | Wallet & payment operations |

---

## What Changed from the Monolith

| Aspect | Monolith | Microservices |
| :--- | :--- | :--- |
| **Wishlist** | In-process controller + shared DB | Independent gRPC service + own DB |
| **Wallet** | In-process controller + shared DB | Independent gRPC service + own DB + own Razorpay |
| **Communication** | Direct function calls | gRPC with protobuf serialization |
| **Auth (inter-service)** | N/A (same process) | JWT passed via gRPC metadata headers |
| **Databases** | 1 shared PostgreSQL | 3 separate PostgreSQL databases |
| **Deployment** | 1 binary | 4 independent processes |
| **Entry Point** | Direct to `:8080` | API Gateway on `:3000` → proxy to `:8080` |
| **Template Loading** | `views/**/*.html` from monolith | Monolith loads its own + wishlist-service templates via `filepath.Glob` |

---

## Roadmap

- [ ] Extract **Auth Service** (JWT issuance, OTP, OAuth) into standalone gRPC service
- [ ] Extract **Catalog Service** (Products, Categories, Offers) into standalone gRPC service
- [ ] Extract **Order Service** (Cart, Checkout, Orders) into standalone gRPC service
- [ ] Add **Docker Compose** for multi-service orchestration
- [ ] Add **service discovery** (Consul or environment-based)
- [ ] Introduce **async messaging** (RabbitMQ/Kafka) for events like order placed, payment confirmed
- [ ] Kubernetes manifests per service with independent scaling

---

## Related Repository

| Repository | Architecture | Description |
| :--- | :--- | :--- |
| [vogueluxe-monolith](https://github.com/Mohammad-Niyas/VogueLuxe-Ecommerce-Monolithic) | Monolithic | Original single-binary version with all domains in one process |
| **vogueluxe-microservices** (this repo) | Microservices | Decomposed version with gRPC service extraction |

---

*Built by [Mohammad Niyas](https://github.com/Mohammad-Niyas)*
