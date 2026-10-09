from fastapi import FastAPI
from fastapi.testclient import TestClient
from langgraph.graph import StateGraph, START, END
from typing import TypedDict
from agenthealth_ref import router
from langgraph_ref import langgraph_router


def snapshot():
    return dict(name='reference', live=True, ready=True,
                capabilities=['health'], dependencies=[])


def test_passive_and_safe_task_contract():
    async def probe(task):
        return {'completed': True, 'success': task['text'] == 'health'}
    app = FastAPI()
    app.include_router(router(snapshot, probe))
    with TestClient(app) as client:
        assert client.head('/health').status_code == 200
        assert client.get('/health').json() == {**snapshot(), 'version': 'v1'}
        assert client.post('/health', json={'safe': True, 'text': 'health'}).json() == {
            'completed': True, 'success': True}
        assert client.post('/health', json={'safe': False, 'text': 'health'}).status_code == 400
        assert client.post('/health', content='x' * 65537).status_code == 413
        assert client.post('/health', content='{} {}').status_code == 400


def test_passive_only_and_no_exception_disclosure():
    app = FastAPI()
    app.include_router(router(snapshot))
    with TestClient(app) as client:
        assert client.post('/health', json={'safe': True, 'text': 'health'}).status_code == 405
    async def broken(_task):
        raise RuntimeError('backend-secret')
    app = FastAPI()
    app.include_router(router(snapshot, broken))
    with TestClient(app) as client:
        response = client.post('/health', json={'safe': True, 'text': 'health'})
        assert response.status_code == 503
        assert 'backend-secret' not in response.text


def test_real_compiled_langgraph():
    class State(TypedDict):
        health_probe: str
        healthy: bool
    async def safe_node(state):
        return {'healthy': state['health_probe'] == 'health'}
    graph = StateGraph(State)
    graph.add_node('safe_health', safe_node)
    graph.add_edge(START, 'safe_health')
    graph.add_edge('safe_health', END)
    app = FastAPI()
    app.include_router(langgraph_router(graph.compile(), snapshot))
    with TestClient(app) as client:
        assert client.post('/health', json={'safe': True, 'text': 'health'}).json()['success'] is True
        assert client.post('/health', json={'safe': True, 'text': 'health', 'downstream': 'peer'}).json()['success'] is False
