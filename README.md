# Scrapp'd 📸✨

> Transform everyday moments into beautiful digital scrapbooks

Scrapp'd is a mobile-first digital scrapbooking platform that helps users find beauty in mundane everyday objects. Photograph receipts, tickets, packaging, and more—our AI removes backgrounds and lets you arrange items on customizable journal pages to create stunning shareable memories.

## 🎯 Project Overview

**Target Audience:** Teens to early 30s, content creators, influencers, travel bloggers
**Core Value:** Aesthetic storytelling through everyday objects
**Monetization:** Freemium model with creator marketplace features

## 🏗️ Architecture

Scrapp'd uses a microservices architecture with the following components:

```
┌─────────────────┐
│  Flutter App    │ ← Mobile Client
└────────┬────────┘
         │
    ┌────▼─────────────────────────────────┐
    │         API Gateway (Go)             │ ← Backend API
    └────┬─────────────────────────────┬───┘
         │                             │
    ┌────▼────────┐            ┌───────▼──────┐
    │ ML Service  │            │  PostgreSQL  │
    │  (Python)   │            │   + Redis    │
    └─────────────┘            └──────────────┘
         │
    ┌────▼────────┐
    │ Cloudflare  │
    │     R2      │ ← Object Storage
    └─────────────┘
```

### Technology Stack

- **Mobile:** Flutter (cross-platform iOS/Android)
- **Backend API:** Go with Gin framework
- **ML Service:** Python with PyTorch (RMBG-2.0/U²-Net)
- **Database:** PostgreSQL + Redis
- **Storage:** Cloudflare R2
- **Infrastructure:** Google Cloud Platform
- **Deployment:** Railway (initial) → GCP (scaling)
- **Payments:** Stripe Philippines
- **Containerization:** Docker

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Flutter SDK (3.16+)
- Go (1.21+)
- Python (3.10+)
- Node.js (for tooling)

### Local Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/scrappd-app.git
cd scrappd-app
```

2. **Set up environment variables**
```bash
# Copy environment templates
cp .env.example .env
cp services/api/.env.example services/api/.env
cp services/ml-service/.env.example services/ml-service/.env
cp mobile/.env.example mobile/.env
```

3. **Start infrastructure services**
```bash
docker-compose up -d postgres redis
```

4. **Run database migrations**
```bash
cd services/api
make migrate-up
```

5. **Start backend services**
```bash
# Terminal 1: API Service
cd services/api
make dev

# Terminal 2: ML Service
cd services/ml-service
make dev
```

6. **Run mobile app**
```bash
cd mobile
flutter pub get
flutter run
```

## 📁 Project Structure

```
scrappd-app/
├── mobile/                    # Flutter mobile application
│   ├── lib/
│   │   ├── core/             # Core utilities, constants
│   │   ├── features/         # Feature-based modules
│   │   ├── shared/           # Shared widgets, services
│   │   └── main.dart
│   ├── assets/               # Images, fonts, etc.
│   └── pubspec.yaml
│
├── services/
│   ├── api/                  # Go backend API
│   │   ├── cmd/              # Application entry points
│   │   ├── internal/         # Private application code
│   │   ├── pkg/              # Public libraries
│   │   ├── migrations/       # Database migrations
│   │   └── Dockerfile
│   │
│   └── ml-service/           # Python ML service
│       ├── src/
│       │   ├── models/       # ML models
│       │   ├── api/          # FastAPI endpoints
│       │   └── utils/        # Utilities
│       ├── requirements.txt
│       └── Dockerfile
│
├── infrastructure/           # IaC and deployment configs
│   ├── terraform/           # GCP infrastructure
│   ├── kubernetes/          # K8s manifests
│   └── railway/             # Railway configs
│
├── docs/                    # Documentation
│   ├── api/                 # API documentation
│   ├── architecture/        # Architecture docs
│   └── guides/              # Development guides
│
├── scripts/                 # Utility scripts
│   ├── setup.sh            # Setup script
│   └── seed-db.sh          # Database seeding
│
├── docker-compose.yml       # Local development
├── .env.example            # Environment template
└── README.md
```

## 🔧 Development

### Running Tests

```bash
# API tests
cd services/api
make test

# ML service tests
cd services/ml-service
pytest

# Mobile tests
cd mobile
flutter test
```

### Code Quality

```bash
# API linting
cd services/api
make lint

# ML service linting
cd services/ml-service
make lint

# Mobile analysis
cd mobile
flutter analyze
```

## 🗺️ Product Roadmap

### Phase 1: MVP (Months 1-3)
- [ ] User authentication & profiles
- [ ] Image upload & background removal
- [ ] Basic canvas editor
- [ ] Simple journal creation
- [ ] Export to image

### Phase 2: Core Features (Months 4-6)
- [ ] Advanced editing tools
- [ ] Template marketplace
- [ ] Social sharing
- [ ] Collections & organization
- [ ] Basic monetization

### Phase 3: Growth (Months 7-9)
- [ ] Collaboration features
- [ ] Creator tools
- [ ] Advanced templates
- [ ] Community features
- [ ] Analytics dashboard

### Phase 4: Scale (Months 10-12)
- [ ] API for third-party integrations
- [ ] Advanced AI features
- [ ] Enterprise features
- [ ] Global expansion

## 💰 Monetization Strategy

### Freemium Model
- **Free Tier:** 10 scrapbooks/month, basic templates
- **Pro ($4.99/month):** Unlimited scrapbooks, premium templates
- **Creator ($9.99/month):** Marketplace access, analytics, collaboration

### Revenue Streams
1. Subscription tiers
2. Template marketplace (70/30 split with creators)
3. Print-on-demand partnerships
4. Brand partnerships for sponsored templates

## 🎯 Go-to-Market Strategy

1. **Launch Phase**
   - Instagram & TikTok content marketing
   - Influencer partnerships (micro-influencers 10k-100k)
   - Travel blogger outreach
   
2. **Growth Phase**
   - User-generated content campaigns
   - Template creator program
   - Community building
   
3. **Scale Phase**
   - Brand partnerships
   - International expansion
   - API ecosystem

## 📊 Success Metrics

- **Acquisition:** 10k users in first 3 months
- **Activation:** 60% create first scrapbook
- **Retention:** 40% weekly active users
- **Revenue:** 5% conversion to paid
- **Referral:** 20% share at least one scrapbook

## 🤝 Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development guidelines.

## 📝 License

Copyright © 2025 Scrapp'd. All rights reserved.

## 📧 Contact

- **Email:** hello@scrappd.app
- **Twitter:** @scrappd_app
- **Instagram:** @scrappd.app

---

Built with ❤️ in the Philippines