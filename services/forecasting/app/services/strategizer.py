"""
Strategy generator - generates actionable strategies per month
"""

from typing import List, Dict, Any
from app.models.forecast import MonthlyForecast


class Strategizer:
    """
    Generates monthly strategies based on forecasts
    """
    
    def generate_strategies(
        self,
        product_id: str,
        forecasts: List[MonthlyForecast],
        product_info: Optional[Dict[str, Any]] = None
    ) -> List[Dict[str, Any]]:
        """
        Generate strategies for each month
        """
        strategies = []
        
        for forecast in forecasts:
            strategy = self._generate_monthly_strategy(forecast, product_info)
            strategies.append({
                "product_id": product_id,
                "month": forecast.month,
                "forecast_id": None,  # Will be set when linked
                "strategy_text": strategy["text"],
                "actions": strategy["actions"],
                "priority": strategy["priority"],
                "estimated_impact": strategy["estimated_impact"]
            })
        
        return strategies
    
    def _generate_monthly_strategy(
        self,
        forecast: MonthlyForecast,
        product_info: Optional[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """
        Generate strategy for a single month
        """
        predicted = forecast.predicted_quantity
        confidence = forecast.confidence_score
        
        # Determine strategy based on forecast
        if predicted == 0:
            return self._no_demand_strategy(forecast.month)
        elif predicted < 50:
            return self._low_demand_strategy(forecast, product_info)
        elif predicted < 200:
            return self._medium_demand_strategy(forecast, product_info)
        else:
            return self._high_demand_strategy(forecast, product_info)
    
    def _no_demand_strategy(self, month: int) -> Dict[str, Any]:
        """Strategy for no/low demand"""
        return {
            "text": f"Bulan {month}: Proyeksi permintaan sangat rendah. Pertimbangkan untuk mengurangi stok atau melakukan promosi untuk meningkatkan penjualan.",
            "actions": {
                "pricing": {
                    "action": "reduce",
                    "percentage": 10,
                    "reason": "Meningkatkan daya tarik produk dengan harga lebih kompetitif"
                },
                "inventory": {
                    "action": "reduce",
                    "quantity": 0,
                    "reason": "Menghindari overstock"
                },
                "marketing": {
                    "channels": ["social", "email"],
                    "budget": 200000,
                    "reason": "Meningkatkan awareness produk"
                }
            },
            "priority": "low",
            "estimated_impact": {
                "sales_increase": "5-10%",
                "cost_reduction": "10-15%"
            }
        }
    
    def _low_demand_strategy(self, forecast: MonthlyForecast, product_info: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """Strategy for low demand"""
        return {
            "text": f"Bulan {forecast.month}: Proyeksi permintaan rendah ({forecast.predicted_quantity} unit). Fokus pada optimasi stok dan pemasaran bertarget.",
            "actions": {
                "pricing": {
                    "action": "maintain",
                    "percentage": 0,
                    "reason": "Mempertahankan harga saat ini"
                },
                "inventory": {
                    "action": "optimize",
                    "quantity": forecast.predicted_quantity,
                    "reason": f"Menjaga stok sesuai proyeksi ({forecast.predicted_quantity} unit)"
                },
                "marketing": {
                    "channels": ["social"],
                    "budget": 300000,
                    "reason": "Meningkatkan visibilitas produk"
                }
            },
            "priority": "medium",
            "estimated_impact": {
                "sales_increase": "10-15%",
                "inventory_optimization": "20-30%"
            }
        }
    
    def _medium_demand_strategy(self, forecast: MonthlyForecast, product_info: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """Strategy for medium demand"""
        return {
            "text": f"Bulan {forecast.month}: Proyeksi permintaan sedang ({forecast.predicted_quantity} unit). Pertahankan operasi normal dengan monitoring ketat.",
            "actions": {
                "pricing": {
                    "action": "maintain",
                    "percentage": 0,
                    "reason": "Harga sudah optimal"
                },
                "inventory": {
                    "action": "restock",
                    "quantity": forecast.predicted_quantity + 20,  # Safety stock
                    "reason": f"Memastikan ketersediaan dengan safety stock"
                },
                "marketing": {
                    "channels": ["social", "email"],
                    "budget": 500000,
                    "reason": "Mempertahankan momentum penjualan"
                }
            },
            "priority": "medium",
            "estimated_impact": {
                "sales_stability": "±5%",
                "customer_satisfaction": "High"
            }
        }
    
    def _high_demand_strategy(self, forecast: MonthlyForecast, product_info: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """Strategy for high demand"""
        return {
            "text": f"Bulan {forecast.month}: Proyeksi permintaan tinggi ({forecast.predicted_quantity} unit). Siapkan stok yang cukup dan pertimbangkan peningkatan kapasitas.",
            "actions": {
                "pricing": {
                    "action": "increase",
                    "percentage": 5,
                    "reason": "Mengoptimalkan margin pada permintaan tinggi"
                },
                "inventory": {
                    "action": "restock",
                    "quantity": forecast.predicted_quantity + 50,  # Higher safety stock
                    "reason": f"Mencegah stockout dengan stok yang memadai"
                },
                "marketing": {
                    "channels": ["social", "email", "marketplace"],
                    "budget": 1000000,
                    "reason": "Maksimalkan penjualan pada periode permintaan tinggi"
                }
            },
            "priority": "high",
            "estimated_impact": {
                "sales_increase": "20-30%",
                "revenue_increase": "25-35%"
            }
        }

