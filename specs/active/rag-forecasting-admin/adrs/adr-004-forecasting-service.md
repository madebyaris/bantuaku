# ADR-004: Use Python Microservice for Forecasting

## Status
Accepted

## Context

We need a 12-month forecasting service with advanced ML capabilities. Options:

1. **Python microservice** (FastAPI) - Separate service
2. **Go module** - Integrated into backend
3. **External service** (third-party API) - Managed service

## Decision

Use **Python microservice (FastAPI)** for forecasting, with Go adapter in backend.

## Rationale

### Advantages

1. **ML Ecosystem**: Python has best ML libraries (scikit-learn, Prophet, statsmodels)
2. **Data Science Tools**: pandas, numpy, matplotlib for data processing
3. **Separation of Concerns**: ML logic isolated from API logic
4. **Independent Scaling**: Scale forecasting service separately
5. **Team Expertise**: Data scientists can work in Python
6. **Future Flexibility**: Easy to add more ML features

### Trade-offs

1. **Service Complexity**: Additional service to deploy and maintain
   - **Mitigation**: Docker Compose makes this straightforward
2. **Network Latency**: HTTP calls between services
   - **Mitigation**: Async jobs, not request-time critical
3. **Language Mix**: Two languages in codebase
   - **Mitigation**: Clear service boundaries, adapter pattern

## Implementation

### Service Structure

```
services/forecasting/
├── app/
│   ├── main.py              # FastAPI app
│   ├── models/
│   │   ├── forecast.py       # Forecast models
│   │   └── strategy.py       # Strategy generation
│   ├── services/
│   │   ├── predictor.py      # Time-series forecasting
│   │   └── strategizer.py    # Strategy generation
│   └── api/
│       └── routes.py         # API endpoints
├── requirements.txt
├── Dockerfile
└── README.md
```

### FastAPI Service

```python
# services/forecasting/app/main.py
from fastapi import FastAPI
from app.api.routes import router

app = FastAPI(title="Forecasting Service")
app.include_router(router)

@app.get("/health")
async def health():
    return {"status": "ok"}
```

### Forecasting Endpoint

```python
# services/forecasting/app/api/routes.py
from fastapi import APIRouter
from app.models.forecast import ForecastRequest, ForecastResponse
from app.services.predictor import ForecastPredictor

router = APIRouter()

@router.post("/forecast", response_model=ForecastResponse)
async def generate_forecast(request: ForecastRequest):
    predictor = ForecastPredictor()
    forecast = await predictor.predict(
        sales_history=request.sales_history,
        trends=request.trends,
        regulation_flags=request.regulation_flags
    )
    return forecast
```

### Go Adapter

```go
// backend/services/forecast/adapter.go
package forecast

type Adapter struct {
    client *http.Client
    baseURL string
}

func (a *Adapter) GenerateForecast(ctx context.Context, inputs ForecastInputs) (*Forecast, error) {
    reqBody := ForecastRequest{
        SalesHistory: inputs.SalesHistory,
        Trends: inputs.Trends,
        RegulationFlags: inputs.RegulationFlags,
    }
    
    resp, err := a.client.Post(a.baseURL+"/forecast", "application/json", reqBody)
    // ... handle response
}
```

### Docker Compose Integration

```yaml
services:
  forecasting:
    build:
      context: ./services/forecasting
      dockerfile: Dockerfile
    ports:
      - "8001:8000"
    environment:
      - DATABASE_URL=${DATABASE_URL}
    depends_on:
      - db
```

## Alternatives Considered

### Go Module (Integrated)

**Pros:**
- Single language codebase
- No network calls
- Simpler deployment

**Cons:**
- Limited ML libraries (no Prophet, limited statsmodels)
- Less mature ML ecosystem
- Harder for data scientists to contribute

**Decision**: Not chosen - ML capabilities insufficient in Go.

### External Service (Third-Party)

**Pros:**
- No ML infrastructure to manage
- Managed service

**Cons:**
- Vendor lock-in
- Cost at scale
- Less control over algorithms
- Data privacy concerns

**Decision**: Not chosen - need control over forecasting logic.

## Consequences

### Positive

- Best ML capabilities (Python ecosystem)
- Independent scaling
- Clear service boundaries
- Easy to extend with more ML features

### Negative

- Additional service to deploy
- Network latency (mitigated with async jobs)
- Two languages in codebase (acceptable trade-off)

### Mitigations

- Use Docker Compose for easy local development
- Implement adapter pattern for clean Go integration
- Use async job queue for forecasting (not request-time)
- Document service API clearly

## Future Considerations

1. **Caching**: Cache forecast results in Redis (1h TTL)
2. **Batch Processing**: Process multiple products in parallel
3. **Model Versioning**: Track which model version generated forecast
4. **A/B Testing**: Compare different forecasting algorithms
5. **Monitoring**: Track forecast accuracy over time

## References

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Prophet Forecasting](https://facebook.github.io/prophet/)
- [scikit-learn](https://scikit-learn.org/)
- [statsmodels](https://www.statsmodels.org/)

