from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def read_root():
    return {"message": "Root"}


@app.get("/items/{item_id}")
def read_item(item_id: int, q: str | None = None):
    return {"item_id": item_id, "query": q}


if __name__ == "__main__":
    print("Run api via \npython -m uvicorn api:app --reload\n\n")
