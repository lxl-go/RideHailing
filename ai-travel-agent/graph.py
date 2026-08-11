import json
import re
from typing import Any, Dict, TypedDict

from amap_client import AmapTravelContextClient
from deepseek_client import DeepSeekClient
from schemas import TravelAdviceRequest, TravelAdviceResponse

try:
    from langgraph.graph import END, StateGraph
except Exception:
    END = None
    StateGraph = None


class TravelAdviceState(TypedDict, total=False):
    request: TravelAdviceRequest
    prompt: str
    raw_model_output: str
    response: TravelAdviceResponse
    client: Any
    amap_client: Any
    amap_context: Dict[str, Any]


def build_prompt(req: TravelAdviceRequest, amap_context: Dict[str, Any] | None = None) -> str:
    mode = normalize_mode(req.mode)
    if mode == "nearby":
        return build_nearby_prompt(req, amap_context or {})
    return (
        "Generate one driver-facing AI travel reminder for a ride-hailing order.\n"
        "Return JSON with exactly these fields: mode, riskLevel, summary, weatherAdvice, routeAdvice, safetyAdvice, trafficAdvice, dispatchAdvice, nearbyTraffic, displayText.\n"
        "All user-facing JSON values must be concise Simplified Chinese.\n"
        "riskLevel must be low, medium, or high. Each advice field must be an array of short strings.\n"
        "mode must be order.\n"
        f"Order ID: {req.orderId}\n"
        f"Driver ID: {req.driverId}\n"
        f"Order status: {req.orderStatus}\n"
        f"Start: {req.startAddress}\n"
        f"End: {req.endAddress}\n"
        f"Driver location: {req.driverLat}, {req.driverLng}\n"
        f"Route distance km: {req.routeDistanceKm}\n"
        f"Estimated minutes: {req.estimatedMinutes}\n"
        f"Known weather text: {req.weatherText}\n"
        f"Scene: {req.scene}\n"
        "Focus on bad weather, route safety, ETA impact, and safe driving. "
        "Do not change the order status or make decisions for the driver."
    )


def build_nearby_prompt(req: TravelAdviceRequest, amap_context: Dict[str, Any]) -> str:
    weather = amap_context.get("weather") or {}
    traffic = amap_context.get("traffic") or {}
    return (
        "Generate one driver-facing AI warning for an idle ride-hailing driver.\n"
        "Return JSON with exactly these fields: mode, riskLevel, summary, weatherAdvice, routeAdvice, safetyAdvice, trafficAdvice, dispatchAdvice, nearbyTraffic, displayText.\n"
        "All user-facing JSON values must be concise Simplified Chinese.\n"
        "riskLevel must be low, medium, or high. Each advice field must be an array of short strings.\n"
        "mode must be nearby.\n"
        f"Driver ID: {req.driverId}\n"
        f"Driver location: {req.driverLat}, {req.driverLng}\n"
        f"Warning radius: 周边5公里, actual radius 4999 meters.\n"
        f"Scene: {req.scene}\n"
        "Amap weather context:\n"
        f"{json.dumps(weather, ensure_ascii=False)}\n"
        "Amap traffic context:\n"
        f"{json.dumps(traffic, ensure_ascii=False)}\n"
        "Focus on nearby congestion, construction or abnormal road conditions, weather risk, and whether the driver should keep taking orders. "
        "Do not invent exact road names when Amap did not return them."
    )


def normalize_mode(mode: str) -> str:
    return "nearby" if str(mode or "").strip().lower() == "nearby" else "order"


def normalize_advice(raw: str, req: TravelAdviceRequest | None = None, amap_context: Dict[str, Any] | None = None) -> TravelAdviceResponse:
    text = str(raw or "").strip()
    match = re.search(r"```(?:json)?\s*(.*?)```", text, re.DOTALL | re.IGNORECASE)
    if match:
        text = match.group(1).strip()
    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        payload = {"summary": text, "displayText": text}

    risk = str(payload.get("riskLevel") or payload.get("risk_level") or "medium").lower()
    if risk not in {"low", "medium", "high"}:
        risk = "medium"
    mode = normalize_mode(payload.get("mode") or (req.mode if req else "order"))

    def string_list(name: str) -> list[str]:
        value = payload.get(name) or []
        if isinstance(value, str):
            return [value]
        if isinstance(value, list):
            return [str(item) for item in value if str(item).strip()]
        return []

    summary = str(payload.get("summary") or "").strip()
    display = str(payload.get("displayText") or payload.get("display_text") or summary).strip()
    if not display:
        display = "AI travel advice generated."
    if not summary:
        summary = display

    return TravelAdviceResponse(
        mode=mode,
        riskLevel=risk,
        summary=summary,
        weatherAdvice=string_list("weatherAdvice") or ["Check weather before departure."],
        routeAdvice=string_list("routeAdvice") or ["Reserve extra time for this route."],
        safetyAdvice=string_list("safetyAdvice") or ["Drive carefully and keep distance."],
        trafficAdvice=string_list("trafficAdvice"),
        dispatchAdvice=string_list("dispatchAdvice"),
        nearbyTraffic=payload.get("nearbyTraffic") if isinstance(payload.get("nearbyTraffic"), dict) else (amap_context or {}),
        displayText=display,
    )


def _prepare_prompt(state: TravelAdviceState) -> TravelAdviceState:
    req = state["request"]
    if normalize_mode(req.mode) == "nearby":
        amap_client = state.get("amap_client") or AmapTravelContextClient()
        state["amap_context"] = amap_client.nearby_context(req.driverLat, req.driverLng, radius_m=4999)
    state["prompt"] = build_prompt(req, state.get("amap_context"))
    return state


def _call_model(state: TravelAdviceState) -> TravelAdviceState:
    client = state.get("client") or DeepSeekClient()
    state["raw_model_output"] = client.complete(state["prompt"])
    return state


def _normalize_output(state: TravelAdviceState) -> TravelAdviceState:
    state["response"] = normalize_advice(state.get("raw_model_output", ""), state.get("request"), state.get("amap_context"))
    return state


def _run_sequential(state: TravelAdviceState) -> TravelAdviceState:
    state = _prepare_prompt(state)
    state = _call_model(state)
    return _normalize_output(state)


def _build_graph():
    if StateGraph is None:
        return None
    graph = StateGraph(TravelAdviceState)
    graph.add_node("prepare_prompt", _prepare_prompt)
    graph.add_node("call_model", _call_model)
    graph.add_node("normalize_output", _normalize_output)
    graph.set_entry_point("prepare_prompt")
    graph.add_edge("prepare_prompt", "call_model")
    graph.add_edge("call_model", "normalize_output")
    graph.add_edge("normalize_output", END)
    return graph.compile()


def generate_travel_advice(req: TravelAdviceRequest, client=None, amap_client=None) -> TravelAdviceResponse:
    state: TravelAdviceState = {"request": req, "client": client, "amap_client": amap_client}
    compiled = _build_graph()
    if compiled is None:
        result: Dict[str, Any] = _run_sequential(state)
    else:
        result = compiled.invoke(state)
    return result["response"]
