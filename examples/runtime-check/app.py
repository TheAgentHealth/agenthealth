"""Dedicated read-only health graph. Run with uvicorn app:app --port 8080."""
from typing import TypedDict
from fastapi import FastAPI
from langgraph.graph import StateGraph, START, END
from agenthealth.integrations.langgraph import langgraph_router


class State(TypedDict):
    health_probe: str
    healthy: bool


async def safe_health(state: State):
    return {'healthy': state['health_probe'] == 'health'}


builder = StateGraph(State)
builder.add_node('safe_health', safe_health)
builder.add_edge(START, 'safe_health')
builder.add_edge('safe_health', END)
app = FastAPI()
app.include_router(langgraph_router(builder.compile(), lambda: {
    'name': 'direct-agent', 'live': True, 'ready': True,
    'capabilities': ['health'], 'dependencies': [],
}))
