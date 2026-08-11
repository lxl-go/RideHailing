import json
import os
from pathlib import Path
import urllib.error
import urllib.request


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


class DeepSeekClient:
    def __init__(self, api_key=None, model=None, base_url=None, timeout=20):
        _load_local_env()
        self.api_key = api_key or os.getenv("DEEPSEEK_API_KEY", "")
        self.model = model or os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
        self.base_url = (base_url or os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")).rstrip("/")
        self.timeout = timeout

    def complete(self, prompt):
        if not self.api_key:
            return json.dumps(
                {
                    "mode": "order",
                    "riskLevel": "medium",
                    "summary": "AI 服务已进入兜底模式，可继续生成出行提醒。",
                    "weatherAdvice": ["出发前再次确认天气变化。"],
                    "routeAdvice": ["优先选择主干道并预留通行时间。"],
                    "safetyAdvice": ["保持安全车距，避免急刹急转。"],
                    "trafficAdvice": ["暂无实时路况，按中等风险处理。"],
                    "dispatchAdvice": ["接单前确认当前位置与乘客上车点距离。"],
                    "nearbyTraffic": {},
                    "displayText": "AI预警可用，配置 DEEPSEEK_API_KEY 后将调用实时模型。",
                }
            )

        payload = {
            "model": self.model,
            "messages": [
                {
                    "role": "system",
                    "content": "You are a driver travel safety assistant. Return strict JSON only.",
                },
                {"role": "user", "content": prompt},
            ],
            "temperature": 0.2,
            "stream": False,
        }
        data = json.dumps(payload).encode("utf-8")
        req = urllib.request.Request(
            f"{self.base_url}/chat/completions",
            data=data,
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                body = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="ignore")
            raise RuntimeError(f"DeepSeek API error {exc.code}: {detail}") from exc
        except urllib.error.URLError as exc:
            raise RuntimeError(f"DeepSeek API unavailable: {exc.reason}") from exc

        return body["choices"][0]["message"]["content"]
