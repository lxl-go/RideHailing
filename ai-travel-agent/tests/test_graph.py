import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from graph import generate_travel_advice
from schemas import TravelAdviceRequest


class FakeDeepSeekClient:
    def complete(self, prompt):
        self.prompt = prompt
        return json.dumps(
            {
                "riskLevel": "high",
                "summary": "Heavy rain may slow the trip.",
                "weatherAdvice": ["Watch for heavy rain"],
                "routeAdvice": ["Avoid low-lying roads"],
                "safetyAdvice": ["Keep extra following distance"],
                "displayText": "High rain risk. Drive carefully.",
            }
        )


class FakeAmapClient:
    def nearby_context(self, lat, lng, radius_m=4999):
        self.lat = lat
        self.lng = lng
        self.radius_m = radius_m
        return {
            "weather": {
                "city": "上海市",
                "weather": "小雨",
                "temperature": "28",
                "winddirection": "东南",
                "windpower": "3",
            },
            "traffic": {
                "status": "1",
                "description": "周边道路整体畅通，部分路段缓行",
            },
        }


class TravelAdviceGraphTest(unittest.TestCase):
    def test_generate_travel_advice_normalizes_model_json(self):
        client = FakeDeepSeekClient()
        result = generate_travel_advice(
            TravelAdviceRequest(
                orderId="5001",
                driverId="2001",
                startAddress="Shanghai Station",
                endAddress="Hongqiao Airport",
                driverLat=31.2304,
                driverLng=121.4737,
                scene="before_departure",
            ),
            client=client,
        )

        self.assertEqual("high", result.riskLevel)
        self.assertEqual("High rain risk. Drive carefully.", result.displayText)
        self.assertEqual(["Watch for heavy rain"], result.weatherAdvice)
        self.assertIn("Shanghai Station", client.prompt)
        self.assertIn("Simplified Chinese", client.prompt)

    def test_generate_nearby_warning_includes_amap_context(self):
        client = FakeDeepSeekClient()
        amap_client = FakeAmapClient()

        generate_travel_advice(
            TravelAdviceRequest(
                mode="nearby",
                driverId="2001",
                driverLat=31.2304,
                driverLng=121.4737,
                scene="idle_warning",
            ),
            client=client,
            amap_client=amap_client,
        )

        self.assertEqual(31.2304, amap_client.lat)
        self.assertEqual(121.4737, amap_client.lng)
        self.assertEqual(4999, amap_client.radius_m)
        self.assertIn("周边5公里", client.prompt)
        self.assertIn("小雨", client.prompt)
        self.assertIn("整体畅通", client.prompt)


if __name__ == "__main__":
    unittest.main()
