"""
Forecasting predictor using time-series analysis
"""

from datetime import datetime, timedelta
from typing import List, Dict, Any
import math
import statistics

from app.models.forecast import ForecastInputs, ForecastResponse, MonthlyForecast


class ForecastPredictor:
    """
    Predicts 12-month forecasts using time-series analysis
    """
    
    def __init__(self):
        self.algorithm = "ensemble"  # ensemble of methods
        self.model_version = "v1.0.0"
    
    async def predict(self, inputs: ForecastInputs) -> ForecastResponse:
        """
        Generate 12-month forecast
        """
        # Parse sales history
        sales_data = self._parse_sales_history(inputs.sales_history)
        
        if len(sales_data) < 7:
            # Insufficient data - return conservative forecast
            return self._generate_conservative_forecast(inputs.product_id)
        
        # Calculate base forecast using time-series methods
        base_forecasts = self._calculate_base_forecast(sales_data)
        
        # Apply exogenous signals
        adjusted_forecasts = self._apply_exogenous_signals(
            base_forecasts,
            inputs.trends_data,
            inputs.regulation_flags,
            inputs.exogenous_factors
        )
        
        # Generate monthly forecasts
        monthly_forecasts = []
        for month in range(1, 13):
            forecast = adjusted_forecasts[month - 1]
            
            # Calculate confidence intervals
            confidence_lower = int(forecast * 0.8)  # 20% lower bound
            confidence_upper = int(forecast * 1.2)  # 20% upper bound
            
            # Calculate confidence score based on data quality
            confidence_score = self._calculate_confidence_score(sales_data, month)
            
            monthly_forecasts.append(MonthlyForecast(
                month=month,
                predicted_quantity=int(forecast),
                confidence_lower=confidence_lower,
                confidence_upper=confidence_upper,
                confidence_score=confidence_score
            ))
        
        return ForecastResponse(
            product_id=inputs.product_id,
            forecast_date=datetime.now().isoformat(),
            algorithm=self.algorithm,
            model_version=self.model_version,
            forecasts=monthly_forecasts,
            metadata={
                "data_points": len(sales_data),
                "trends_signals": len(inputs.trends_data) if inputs.trends_data else 0,
                "regulation_flags": len(inputs.regulation_flags) if inputs.regulation_flags else 0,
            }
        )
    
    def _parse_sales_history(self, sales_history: List) -> List[float]:
        """Parse sales history into numeric array"""
        return [point.quantity for point in sales_history]
    
    def _calculate_base_forecast(self, sales_data: List[float]) -> List[float]:
        """
        Calculate base forecast using ensemble of methods
        Returns 12 monthly predictions
        """
        # Simple Moving Average (30-day)
        sma = self._simple_moving_average(sales_data, min(30, len(sales_data)))
        
        # Exponential Smoothing
        es = self._exponential_smoothing(sales_data, alpha=0.3)
        
        # Trend projection
        trend = self._trend_projection(sales_data)
        
        # Seasonal adjustment (if enough data)
        seasonal = self._seasonal_adjustment(sales_data)
        
        # Ensemble: weighted average
        base_daily = sma * 0.3 + es * 0.3 + trend * 0.25 + seasonal * 0.15
        
        # Project to monthly (average days per month)
        days_per_month = 30.44
        monthly_base = base_daily * days_per_month
        
        # Generate 12 months with trend decay
        forecasts = []
        for month in range(1, 13):
            # Apply slight decay for longer horizons
            decay_factor = 1.0 - (month - 1) * 0.02  # 2% decay per month
            forecasts.append(monthly_base * decay_factor)
        
        return forecasts
    
    def _simple_moving_average(self, data: List[float], period: int) -> float:
        """Calculate simple moving average"""
        if len(data) < period:
            period = len(data)
        if period == 0:
            return 0.0
        return sum(data[-period:]) / period
    
    def _exponential_smoothing(self, data: List[float], alpha: float) -> float:
        """Calculate exponential smoothing"""
        if len(data) == 0:
            return 0.0
        result = data[0]
        for value in data[1:]:
            result = alpha * value + (1 - alpha) * result
        return result
    
    def _trend_projection(self, data: List[float]) -> float:
        """Project trend forward"""
        if len(data) < 2:
            return data[0] if len(data) > 0 else 0.0
        
        # Linear regression
        n = len(data)
        x_mean = (n - 1) / 2
        y_mean = sum(data) / n
        
        numerator = sum((i - x_mean) * (data[i] - y_mean) for i in range(n))
        denominator = sum((i - x_mean) ** 2 for i in range(n))
        
        if denominator == 0:
            return y_mean
        
        slope = numerator / denominator
        intercept = y_mean - slope * x_mean
        
        # Project 7 days forward
        return intercept + slope * (n + 6)
    
    def _seasonal_adjustment(self, data: List[float]) -> float:
        """Calculate seasonal adjustment"""
        if len(data) < 30:
            return sum(data) / len(data) if len(data) > 0 else 0.0
        
        # Simple seasonal: average of last 30 days
        return sum(data[-30:]) / 30
    
    def _apply_exogenous_signals(
        self,
        base_forecasts: List[float],
        trends_data: Optional[List],
        regulation_flags: Optional[List],
        exogenous_factors: Optional[Dict[str, Any]]
    ) -> List[float]:
        """
        Adjust forecasts based on exogenous signals
        """
        adjusted = base_forecasts.copy()
        
        # Apply trends signals
        if trends_data:
            trend_impact = self._calculate_trend_impact(trends_data)
            for i in range(len(adjusted)):
                # Trend impact decays over time
                impact_factor = 1.0 + (trend_impact * (1.0 - i * 0.05))
                adjusted[i] *= impact_factor
        
        # Apply regulation flags
        if regulation_flags:
            regulation_impact = self._calculate_regulation_impact(regulation_flags)
            for i in range(len(adjusted)):
                adjusted[i] *= (1.0 + regulation_impact)
        
        return adjusted
    
    def _calculate_trend_impact(self, trends_data: List) -> float:
        """Calculate impact of trends signals"""
        if not trends_data:
            return 0.0
        
        # Average trend value (normalized)
        avg_trend = sum(t.value for t in trends_data) / len(trends_data)
        
        # Normalize to -0.1 to +0.1 impact (10% max change)
        return (avg_trend - 50) / 500.0
    
    def _calculate_regulation_impact(self, regulation_flags: List) -> float:
        """Calculate impact of regulation flags"""
        if not regulation_flags:
            return 0.0
        
        total_impact = 0.0
        for flag in regulation_flags:
            score = flag.relevance_score
            if flag.impact == "positive":
                total_impact += score * 0.05  # Up to 5% increase
            elif flag.impact == "negative":
                total_impact -= score * 0.05  # Up to 5% decrease
        
        return total_impact / len(regulation_flags)
    
    def _calculate_confidence_score(self, sales_data: List[float], month: int) -> float:
        """Calculate confidence score for a forecast month"""
        if len(sales_data) < 7:
            return 0.3
        
        # Base confidence from data quality
        data_quality = min(1.0, len(sales_data) / 90.0)  # More data = higher confidence
        
        # Decay confidence for longer horizons
        horizon_decay = 1.0 - (month - 1) * 0.05  # 5% decay per month
        
        # Variability penalty
        if len(sales_data) > 1:
            variance = statistics.variance(sales_data)
            mean = statistics.mean(sales_data)
            cv = math.sqrt(variance) / mean if mean > 0 else 1.0
            variability_penalty = min(1.0, 1.0 - cv * 0.5)
        else:
            variability_penalty = 0.5
        
        confidence = data_quality * horizon_decay * variability_penalty
        return max(0.1, min(1.0, confidence))
    
    def _generate_conservative_forecast(self, product_id: str) -> ForecastResponse:
        """Generate conservative forecast when data is insufficient"""
        monthly_forecasts = []
        for month in range(1, 13):
            monthly_forecasts.append(MonthlyForecast(
                month=month,
                predicted_quantity=0,
                confidence_lower=0,
                confidence_upper=0,
                confidence_score=0.1
            ))
        
        return ForecastResponse(
            product_id=product_id,
            forecast_date=datetime.now().isoformat(),
            algorithm="conservative",
            model_version=self.model_version,
            forecasts=monthly_forecasts,
            metadata={"insufficient_data": True}
        )

