# Bantuaku SaaS

**AI-Chat-First Forecasting Assistant untuk UMKM Indonesia**

![Bantuaku](https://img.shields.io/badge/Status-Hackathon%20MVP-purple)
![Go](https://img.shields.io/badge/Backend-Go%201.25-00ADD8)
![React](https://img.shields.io/badge/Frontend-React%2018-61DAFB)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL%2018-336791)
![Kolosal.ai](https://img.shields.io/badge/AI-Kolosal.ai-blue)

## 🎯 Overview

Bantuaku adalah platform SaaS yang membantu UMKM Indonesia membuat keputusan bisnis berbasis data melalui **AI chat sebagai interface utama**. Platform ini mengumpulkan informasi bisnis secara conversational dan menghasilkan insights praktis.

### ✨ Fitur Utama

- **🤖 AI Assistant Chat** - Interface utama untuk mengumpulkan data bisnis secara conversational (powered by Kolosal.ai)
- **📊 Forecast** - Proyeksi penjualan 30/60/90 hari ke depan berdasarkan data penjualan yang diinput user
- **🌍 Market Prediction** - Prediksi tren pasar lokal (Indonesia) dan global untuk produk Anda
- **📢 Marketing Recommendation** - Rekomendasi kampanye marketing dan strategi promosi
- **⚖️ Government Regulation** - Informasi peraturan pemerintah Indonesia yang relevan dengan bisnis
- **📁 File Upload** - Upload CSV, XLSX, atau PDF untuk ekstraksi data otomatis (OCR powered by Kolosal.ai)

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
├── backend/                    # Go backend API
│   ├── config/                 # Configuration
│   ├── handlers/               # HTTP handlers
│   │   ├── chat.go            # Chat & conversation handlers
│   │   ├── files.go           # File upload handlers
│   │   ├── insights.go        # Insights generation handlers
│   │   └── companies.go       # Company profile handlers
│   ├── middleware/             # Auth, CORS, logging
│   ├── models/                 # Data models
│   │   ├── company.go         # Company & CompanyProfile
│   │   ├── conversation.go    # Conversation & Message
│   │   ├── file_upload.go     # FileUpload & ExtractedData
│   │   ├── data_source.go     # DataSource
│   │   └── insight.go         # Insight & result types
│   ├── services/               # Business logic
│   │   ├── kolosal/           # Kolosal.ai client (Chat & OCR)
│   │   ├── storage/            # Database & Redis
│   │   ├── chat/               # Chat service (TODO)
│   │   ├── ingestion/         # File processing service (TODO)
│   │   ├── forecasting/       # Forecast service (TODO)
│   │   ├── connectors/        # External data connectors (TODO)
│   │   └── insights/          # Insights generation (TODO)
│   └── main.go
├── frontend/                   # React frontend
│   ├── src/
│   │   ├── components/        # UI components
│   │   ├── pages/             # Page components
│   │   │   ├── AIChatPage.tsx        # AI Chat interface
│   │   │   ├── ForecastPage.tsx     # Forecast insights
│   │   │   ├── MarketPredictionPage.tsx  # Market predictions
│   │   │   ├── MarketingPage.tsx     # Marketing recommendations
│   │   │   └── RegulationPage.tsx    # Government regulations
│   │   ├── state/             # State management (Zustand)
│   │   └── lib/               # Utilities & API clients
│   └── package.json
├── database/
│   └── migrations/            # SQL migrations
│       └── 003_add_chat_tables.sql  # Chat, ingestion, insights tables
├── .docs-private/              # Product & technical documentation
├── docker-compose.yml
├── .env.example                # Environment variables template
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

### Chat & Conversations (AI-First Interface)
- `POST /api/v1/chat/start` - Start new conversation
- `POST /api/v1/chat/message` - Send message to AI assistant
- `GET /api/v1/chat/conversations` - List all conversations
- `GET /api/v1/chat/messages` - Get messages from a conversation

### File Uploads
- `POST /api/v1/files/upload` - Upload CSV/XLSX/PDF files (with OCR processing)
- `GET /api/v1/files/{id}` - Get file upload information

### Insights (Four Outcome Types)
- `POST /api/v1/insights/forecast` - Generate forecast insights
- `POST /api/v1/insights/market` - Generate market prediction insights
- `POST /api/v1/insights/marketing` - Generate marketing recommendations
- `POST /api/v1/insights/regulation` - Generate government regulation insights
- `GET /api/v1/insights` - Get insight history

### Companies
- `GET /api/v1/companies` - List user's companies
- `GET /api/v1/companies/{id}` - Get company profile (aggregated data)

### Dashboard
- `GET /api/v1/dashboard/summary` - Get dashboard KPIs

### Legacy AI (Deprecated)
- `POST /api/v1/ai/analyze` - Legacy AI analyze endpoint

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

### MVP (Hackathon) - AI-Chat-First Architecture
- ✅ **AI Chat Interface** - Conversational data collection in Bahasa Indonesia
- ✅ **File Upload & OCR** - CSV/XLSX/PDF upload with automatic text extraction (Kolosal.ai OCR)
- ✅ **Forecast Insights** - 30/60/90-day sales forecasting (based on user-provided sales data)
- ✅ **Market Prediction** - Local (Indonesia) and global market trend analysis
- ✅ **Marketing Recommendations** - AI-generated campaign ideas and strategies
- ✅ **Government Regulations** - Indonesia-specific regulatory information
- ✅ **Company Profile** - Aggregated business data from all sources
- ✅ **Dashboard** - Overview of business KPIs and insights

### Roadmap
- [ ] **External Data Connectors** - Tokopedia, Shopee, Bukalapak marketplace scraping
- [ ] **Google Trends Integration** - Real-time market trend data
- [ ] **Regulation Scraper** - Automated peraturan.go.id monitoring
- [ ] **Advanced Forecasting** - ML-based time-series forecasting
- [ ] **Mobile App** - React Native mobile application
- [ ] **Billing & Subscriptions** - 3-tier pricing (Free, Pro, Enterprise)
- [ ] **Multi-Company Management** - Support for multiple businesses per user

## 🏗️ Architecture

### High-Level Architecture

```
┌─────────────┐     ┌─────────────────────────────────────┐     ┌─────────────┐
│   Frontend  │────▶│           Backend (Go)              │────▶│  PostgreSQL │
│   (React)   │     │  ┌─────────┐  ┌─────────┐         │     │             │
│             │     │  │  Chat   │  │Ingestion│         │     └─────────────┘
│  - AI Chat  │     │  │ Service │  │ Service │         │
│  - Forecast │     │  └─────────┘  └─────────┘         │     ┌─────────────┐
│  - Market   │     │  ┌─────────┐  ┌─────────┐         │────▶│   Redis     │
│  - Marketing│     │  │Forecast │  │Connector│         │     │   (Cache)   │
│  - Regulation│    │  │ Service │  │ Service │         │     └─────────────┘
└─────────────┘     │  └─────────┘  └─────────┘         │
                    │  ┌─────────┐                       │     ┌─────────────┐
                    │  │ Insights│                       │────▶│ Kolosal.ai  │
                    │  │ Service │                       │     │  (Chat+OCR) │
                    │  └─────────┘                       │     └─────────────┘
                    └─────────────────────────────────────┘
```

### Key Components

- **Chat Service** - Handles AI conversations and message history
- **Ingestion Service** - Processes file uploads (CSV/XLSX/PDF) with OCR
- **Forecast Service** - Generates sales forecasts based on user-provided sales data
- **Connector Service** - External data sources (marketplaces, trends, regulations)
- **Insights Service** - Generates four types of insights (forecast, market, marketing, regulation)

## 📝 License

MIT License - see [LICENSE](LICENSE) file.

## 👥 Team

**Enggan Ngoding, Pecut AI**

Built with ❤️ for IMPHNEN x Kolosal.ai Hackathon 2025

### Team Members

- [@madebyaris](https://github.com/madebyaris)
- [@tobangado69](https://github.com/tobangado69)

---

**Bantuaku** - Membantu UMKM Indonesia tumbuh dengan AI dan data 🇮🇩

---

## 🌟 How It Works

1. **Start Chat** - User begins conversation with AI Assistant
2. **Data Collection** - AI asks about company, products, location, business model
3. **File Upload** - User can upload CSV/XLSX/PDF files for automatic data extraction
4. **Profile Building** - System builds comprehensive Company Profile from conversations and files
5. **Generate Insights** - User navigates to four outcome pages:
   - **Forecast** - Sales projections (generated if user provides sales data via chat or file upload)
   - **Market Prediction** - Local and global market trends
   - **Marketing Recommendation** - Campaign ideas and strategies
   - **Government Regulation** - Relevant Indonesian regulations

All powered by **Kolosal.ai** for natural language understanding and OCR processing.
