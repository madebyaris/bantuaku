# Bantuaku SaaS

**AI-Powered Demand Forecasting Platform untuk UMKM Indonesia**

![Bantuaku](https://img.shields.io/badge/Status-Hackathon%20MVP-purple)
![Go](https://img.shields.io/badge/Backend-Go%201.25-00ADD8)
![React](https://img.shields.io/badge/Frontend-React%2018-61DAFB)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL%2018-336791)

## 🎯 Overview

Bantuaku adalah platform SaaS yang membantu UMKM Indonesia membuat keputusan bisnis berbasis data dengan:

- **Forecasting Permintaan** - Prediksi penjualan 30/60/90 hari ke depan berdasarkan data historis
- **Proyeksi Permintaan** - Rekomendasi proyeksi permintaan produk berdasarkan tren penjualan
- **Integrasi WooCommerce** - Sinkronisasi produk dan pesanan
- **AI Assistant** - Tanya jawab bisnis dalam Bahasa Indonesia (powered by Kolosal.ai)
- **Sentiment Analysis** - Pantau sentiment pasar dan social media

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Node.js 20+ (untuk development)
- Go 1.25+ (untuk development)

### Running with Docker

```bash
# Clone repository
git clone https://github.com/your-org/bantuaku.git
cd bantuaku

# Start all services
make dev

# Or manually:
docker-compose up --build
```

Aplikasi akan berjalan di:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Demo Account

```
Email: demo@bantuaku.id
Password: demo123
```

## 📁 Project Structure

```
bantuaku/
├── backend/               # Go backend API
│   ├── config/           # Configuration
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Auth, CORS, logging
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   │   └── storage/      # Database & Redis
│   └── main.go
├── frontend/              # React frontend
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── pages/        # Page components
│   │   ├── state/        # State management
│   │   └── lib/          # Utilities & API
│   └── package.json
├── database/
│   └── migrations/       # SQL migrations
├── docker-compose.yml
└── Makefile
```

## 🔧 Development

### Backend (Go)

```bash
cd backend
go mod download
go run main.go
```

### Frontend (React)

```bash
cd frontend
npm install
npm run dev
```

### Environment Variables

Copy the example environment file and configure it:

```bash
# Copy the example file
cp .env.example .env

# Edit .env and add your values (especially KOLOSAL_API_KEY)
# See .env.example for all available configuration options
```

**Required variables:**
- `KOLOSAL_API_KEY` - Get from https://api.kolosal.ai (optional for basic features)
- `JWT_SECRET` - Generate with: `openssl rand -base64 32` (change from default!)

**Quick setup:**
```env
# Minimum required for local development
KOLOSAL_API_KEY=your-api-key-here
JWT_SECRET=your-secure-secret-here
```

See `.env.example` for complete configuration options and documentation.

## 📚 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login

### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/{id}` - Get product
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product

### Sales
- `POST /api/v1/sales/manual` - Record manual sale
- `POST /api/v1/sales/import-csv` - Import CSV
- `GET /api/v1/sales` - List sales history

### Integrations
- `POST /api/v1/integrations/woocommerce/connect` - Connect WooCommerce
- `GET /api/v1/integrations/woocommerce/sync-status` - Get sync status
- `POST /api/v1/integrations/woocommerce/sync-now` - Trigger sync

### Forecasting
- `GET /api/v1/forecasts/{product_id}` - Get product forecast
- `GET /api/v1/recommendations` - Get demand forecast recommendations

### AI
- `POST /api/v1/ai/analyze` - Ask AI assistant

### Dashboard
- `GET /api/v1/dashboard/summary` - Get dashboard KPIs

## 🎨 Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.25 (net/http) |
| Frontend | React 18 + Vite + Tailwind |
| Database | PostgreSQL 18 |
| Cache | Redis 7 |
| AI | Kolosal.ai (Chat & OCR) |
| Deployment | Docker |

## 📊 Features

### MVP (Hackathon)
- ✅ Manual data input (form + CSV)
- ✅ WooCommerce integration
- ✅ 30-day demand forecasting
- ✅ Basic sentiment analysis
- ✅ AI chat in Bahasa Indonesia
- ✅ Dashboard with KPIs

### Roadmap
- [ ] Shopee/Tokopedia integration
- [ ] Mobile app (React Native)
- [ ] Advanced ML forecasting
- [ ] Billing & subscriptions
- [ ] Multi-store management

## 🏗️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Backend   │────▶│  PostgreSQL │
│   (React)   │     │    (Go)     │     │             │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                    ┌─────▼─────┐
                    │   Redis   │
                    │  (Cache)  │
                    └───────────┘
```

## 📝 License

MIT License - see [LICENSE](LICENSE) file.

## 👥 Team

**Enggan Ngoding, Pecut AI**

Built with ❤️ for IMPHNEN x Kolosal.ai Hackathon 2025

### Team Members

- [@madebyaris](https://github.com/madebyaris)
- [@tobangado69](https://github.com/tobangado69)

---

**Bantuaku** - Membantu UMKM Indonesia tumbuh dengan data 🇮🇩
