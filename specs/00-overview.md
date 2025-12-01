# Project Overview: Bantuaku SaaS

## Project Description
AI-Powered Inventory & Demand Forecasting Platform for Indonesian UMKM (Micro, Small & Medium Enterprises). Helps UMKM reduce inventory waste (IDR 3-5M/month) through data-driven forecasting and AI recommendations.

## Project Goals
- Deliver hackathon MVP in 7 days
- Enable UMKM to forecast demand without technical complexity
- Support multiple data input modes (manual, CSV, WooCommerce)
- Provide AI-powered insights in Bahasa Indonesia
- Scale to 5,000+ paying users in Year 1

## Architecture Overview
Built using Vertical Driven Development (VDD) principles:

**Backend:** Go 1.22 (net/http) + PostgreSQL 16 + Redis 7
- 8 vertical slices (auth, products, sales, integrations, forecasts, sentiment, ai, dashboard)
- 20+ REST API endpoints
- JWT authentication, multi-tenant isolation

**Frontend:** React 18 + Vite + Tailwind + shadcn-style components
- 6 main pages (auth, dashboard, products, data-input, integrations, ai-chat)
- Zustand for state management
- Responsive design, Bahasa Indonesia UI

**Infrastructure:** Docker Compose
- PostgreSQL for primary data
- Redis for caching (forecasts, sentiment, AI responses)
- Nginx for frontend serving

## Technology Stack
- **Backend**: Go 1.22, PostgreSQL 16, Redis 7
- **Frontend**: React 18, Vite, Tailwind CSS, TypeScript
- **AI**: OpenAI GPT-4o Mini
- **Deployment**: Docker, Docker Compose
- **Development**: SDD (Spec-Driven Development) workflow

## Current Status
- **Phase**: MVP Complete ✅
- **Version**: 0.1.0 (Hackathon MVP)
- **Last Updated**: 2025-12-01
- **Demo Ready**: Yes

## Active Features
- [feat-002-bantuaku-mvp](active/feat-002-bantuaku-mvp/) - Complete MVP with all 8 vertical slices

## Completed Features
- ✅ Platform Foundation & Environment
- ✅ Auth & Store Onboarding
- ✅ Manual & CSV Sales Data Input
- ✅ WooCommerce Integration
- ✅ Forecasting & Inventory Recommendations (30-day)
- ✅ Sentiment & Market Insights (MVP level)
- ✅ AI Assistant (Bahasa Indonesia)
- ✅ Dashboard & Demo Narrative

## Backlog Features
- Shopee/Tokopedia marketplace integrations
- Mobile app (React Native)
- Billing & subscriptions (Stripe)
- Advanced ML forecasting models
- Legal/regulation intelligence
- Multi-store management (Enterprise tier)

## Team
- **Project**: Bantuaku SaaS
- **Hackathon**: IMPHNEN x Kolosal.ai 2025
- **Timeline**: MVP in 7 days

## Links
- [Feature Index](index.md)
- [Guidelines](../.sdd/guidelines.md)
- [Configuration](../.sdd/config.json)
- [Templates](../.sdd/templates/)
- [PRD](../.docs-private/prd.md) (Private)

## Demo
- **URL**: http://localhost:3000
- **Credentials**: demo@bantuaku.id / demo123
- **Start**: `make dev`
