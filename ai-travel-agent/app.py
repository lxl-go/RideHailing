from fastapi import FastAPI

from graph import generate_travel_advice
from schemas import TravelAdviceRequest

app = FastAPI(title="RideHailing AI Travel Agent")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/travel-advice")
def travel_advice(req: dict):
    return generate_travel_advice(TravelAdviceRequest.from_mapping(req)).to_dict()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("app:app", host="127.0.0.1", port=8011, reload=False)
