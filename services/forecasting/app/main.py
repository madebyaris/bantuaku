"""
Forecasting Service - Python FastAPI microservice
12-month forecasting with exogenous signals
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional, Dict, Any
import uvicorn

app = FastAPI(title="Forecasting Service", version="1.0.0")


class SalesHistoryPoint(BaseModel):
    date: str  # ISO date format
    quantity: int


class TrendSignal(BaseModel):
    keyword: str
    time: str
    value: int


class RegulationFlag(BaseModel):
    regulation_id: str
    relevance_score: float
    impact: str  # 'positive', 'negative', 'neutral'


class ForecastInputs(BaseModel):
    product_id: str
    sales_history: List[SalesHistoryPoint]
    trends_data: Optional[List[TrendSignal]] = None
    regulation_flags: Optional[List[RegulationFlag]] = None
    exogenous_factors: Optional[Dict[str, Any]] = None


class MonthlyForecast(BaseModel):
    month: int  # 1-12
    predicted_quantity: int
    confidence_lower: Optional[int] = None
    confidence_upper: Optional[int] = None
    confidence_score: float


class ForecastResponse(BaseModel):
    product_id: str
    forecast_date: str
    algorithm: str
    model_version: str
    forecasts: List[MonthlyForecast]
    metadata: Optional[Dict[str, Any]] = None


@app.get("/health")
async def health():
    return {"status": "ok", "service": "forecasting"}


@app.post("/forecast", response_model=ForecastResponse)
async def generate_forecast(inputs: ForecastInputs):
    """
    Generate 12-month forecast for a product
    """
    try:
        # Import forecasting logic
        from app.services.predictor import ForecastPredictor
        
        predictor = ForecastPredictor()
        forecast = await predictor.predict(inputs)
        
        return forecast
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Forecast generation failed: {str(e)}")


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)

