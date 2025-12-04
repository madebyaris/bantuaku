"""
Forecast models
"""

from pydantic import BaseModel
from typing import List, Optional, Dict, Any


class SalesHistoryPoint(BaseModel):
    date: str
    quantity: int


class TrendSignal(BaseModel):
    keyword: str
    time: str
    value: int


class RegulationFlag(BaseModel):
    regulation_id: str
    relevance_score: float
    impact: str


class ForecastInputs(BaseModel):
    product_id: str
    sales_history: List[SalesHistoryPoint]
    trends_data: Optional[List[TrendSignal]] = None
    regulation_flags: Optional[List[RegulationFlag]] = None
    exogenous_factors: Optional[Dict[str, Any]] = None


class MonthlyForecast(BaseModel):
    month: int
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

