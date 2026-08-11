from dataclasses import asdict, dataclass, field
from typing import Any, Dict, List


@dataclass
class TravelAdviceRequest:
    mode: str = "order"
    orderId: str = ""
    driverId: str = ""
    orderStatus: str = ""
    startAddress: str = ""
    endAddress: str = ""
    driverLat: float = 0
    driverLng: float = 0
    routeDistanceKm: float = 0
    estimatedMinutes: float = 0
    weatherText: str = ""
    scene: str = "before_departure"
    userRole: str = "driver"

    @classmethod
    def from_mapping(cls, payload: Dict[str, Any]):
        payload = payload or {}
        return cls(
            mode=str(payload.get("mode") or "order"),
            orderId=str(payload.get("orderId") or ""),
            driverId=str(payload.get("driverId") or ""),
            orderStatus=str(payload.get("orderStatus") or payload.get("status") or ""),
            startAddress=str(payload.get("startAddress") or ""),
            endAddress=str(payload.get("endAddress") or ""),
            driverLat=float(payload.get("driverLat") or 0),
            driverLng=float(payload.get("driverLng") or 0),
            routeDistanceKm=float(payload.get("routeDistanceKm") or payload.get("distanceKm") or 0),
            estimatedMinutes=float(payload.get("estimatedMinutes") or payload.get("durationMinutes") or 0),
            weatherText=str(payload.get("weatherText") or ""),
            scene=str(payload.get("scene") or "before_departure"),
            userRole=str(payload.get("userRole") or "driver"),
        )


@dataclass
class TravelAdviceResponse:
    mode: str = "order"
    riskLevel: str = "medium"
    summary: str = ""
    weatherAdvice: List[str] = field(default_factory=list)
    routeAdvice: List[str] = field(default_factory=list)
    safetyAdvice: List[str] = field(default_factory=list)
    trafficAdvice: List[str] = field(default_factory=list)
    dispatchAdvice: List[str] = field(default_factory=list)
    nearbyTraffic: Dict[str, Any] = field(default_factory=dict)
    displayText: str = ""

    def to_dict(self):
        return asdict(self)
