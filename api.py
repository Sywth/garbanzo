from fastapi import FastAPI

# Initialize FastAPI instance
app = FastAPI()


# Root endpoint
@app.get("/")
def read_root():
    return {"message": "Hello, World"}


# Example parameterized endpoint
@app.get("/items/{item_id}")
def read_item(item_id: int, q: str | None = None):
    return {"item_id": item_id, "query": q}
