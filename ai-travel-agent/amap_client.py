import json
import os
from pathlib import Path
from urllib.parse import urlencode
import urllib.error
import urllib.request


AMAP_REST_BASE_URL = "https://restapi.amap.com"


def _load_local_env():
    env_path = Path(__file__).resolve().parent / ".env"
    if not env_path.exists():
        return
    for line in env_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip().strip('"').strip("'"))


class AmapTravelContextClient:
    def __init__(self, key=None, base_url=None, timeout=8):
        _load_local_env()
        self.key = key or os.getenv("AMAP_KEY") or os.getenv("medilive_gd") or ""
        self.base_url = (base_url or os.getenv("AMAP_BASE_URL") or AMAP_REST_BASE_URL).rstrip("/")
        self.timeout = timeout

    def nearby_context(self, lat, lng, radius_m=4999):
        radius = max(1, min(int(radius_m or 4999), 4999))
        context = {
            "radiusMeters": radius,
            "location": {"lat": lat, "lng": lng},
            "amapConfigured": bool(self.key),
        }
        if not self.key:
            context.update(
                {
                    "weather": {"weather": "未知", "temperature": "", "city": ""},
                    "traffic": {"status": "unknown", "description": "未配置 AMAP_KEY，已使用 AI 兜底生成周边预警。"},
                }
            )
            return context

        try:
            regeo = self._get(
                "/v3/geocode/regeo",
                {
                    "location": f"{lng},{lat}",
                    "extensions": "base",
                    "radius": radius,
                },
            )
            adcode = (
                regeo.get("regeocode", {})
                .get("addressComponent", {})
                .get("adcode", "")
            )
            context["regeo"] = regeo.get("regeocode", {})
            if adcode:
                context["weather"] = self._weather(adcode)
            context["traffic"] = self._traffic_circle(lat, lng, radius)
        except Exception as exc:
            context["error"] = str(exc)
            context.setdefault("weather", {"weather": "未知", "temperature": "", "city": ""})
            context.setdefault("traffic", {"status": "unknown", "description": "高德路况暂不可用，已使用 AI 兜底分析。"})
        return context

    def _weather(self, adcode):
        payload = self._get("/v3/weather/weatherInfo", {"city": adcode, "extensions": "base"})
        lives = payload.get("lives") or []
        return lives[0] if lives else {"adcode": adcode}

    def _traffic_circle(self, lat, lng, radius):
        payload = self._get(
            "/v3/traffic/status/circle",
            {
                "location": f"{lng},{lat}",
                "radius": radius,
                "extensions": "all",
            },
        )
        return payload.get("trafficinfo") or payload

    def _get(self, path, params):
        query = dict(params)
        query["key"] = self.key
        url = f"{self.base_url}{path}?{urlencode(query)}"
        with urllib.request.urlopen(url, timeout=self.timeout) as resp:
            payload = json.loads(resp.read().decode("utf-8"))
        if str(payload.get("status")) != "1":
            raise RuntimeError(payload.get("info") or "Amap request failed")
        return payload
