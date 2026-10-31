"""Integration regressions using the installed SDK and actual compiled graphs."""
import asyncio
import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient
from agenthealth.integrations.langgraph import langgraph_router, health_node
from test_reference import snapshot


class Graph:
    async def ainvoke(self, state):
        return {'healthy': state['health_probe'] == 'health'}


def test_selected_peer_is_contacted_and_unknown_peer_rejected():
    calls = []
    async def peer(text):
        calls.append(text)
        return True
    app = FastAPI()
    app.include_router(langgraph_router(Graph(), snapshot, peers={'peer': peer}))
    with TestClient(app) as client:
        result = client.post('/health', json={'safe': True, 'text': 'health', 'downstream': 'peer'})
        assert result.json() == {'completed': True, 'success': True, 'downstream': 'peer'}
        assert calls == ['health']
        result = client.post('/health', json={'safe': True, 'text': 'health', 'downstream': 'other'})
        assert result.json()['success'] is False
        assert calls == ['health']


@pytest.mark.parametrize('mode', ['exception', 'timeout', 'truthy'])
def test_peer_failures_are_bounded_and_redacted(mode):
    async def peer(text):
        if mode == 'exception':
            raise RuntimeError('secret')
        if mode == 'timeout':
            await asyncio.sleep(5)
        return 'yes'
    app = FastAPI()
    app.include_router(langgraph_router(Graph(), snapshot, peers={'peer': peer}))
    with TestClient(app) as client:
        response = client.post('/health', json={'safe': True, 'text': 'health', 'downstream': 'peer'})
        assert 'secret' not in response.text
        if mode == 'truthy':
            assert response.json()['success'] is False
        else:
            assert response.status_code == 503


def test_consumer_preserves_report_errors_and_cancellation():
    report = object()
    config = object()
    class Client:
        async def async_check(self, selected):
            assert selected is config
            return report
    async def run():
        assert await health_node(Client(), config)({'configuration': 'untrusted'}) == {'agenthealth': report}
        class Broken:
            async def async_check(self, selected):
                raise RuntimeError('tool failure')
        with pytest.raises(RuntimeError):
            await health_node(Broken(), config)({})
        class Cancelled:
            async def async_check(self, selected):
                raise asyncio.CancelledError
        with pytest.raises(asyncio.CancelledError):
            await health_node(Cancelled(), config)({})
    asyncio.run(run())


def test_peer_http_exchange_and_compiled_consumer():
    import httpx
    from langgraph.graph import StateGraph, START, END
    from typing import TypedDict
    peer_app = FastAPI()
    peer_app.include_router(langgraph_router(Graph(), snapshot))
    requests = []
    async def exchange(text):
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=peer_app),
                                     base_url='http://peer') as client:
            response = await client.post('/health', json={'safe': True, 'text': text})
            requests.append(response.status_code)
            return response.status_code == 200 and response.json().get('success') is True
    app = FastAPI()
    app.include_router(langgraph_router(Graph(), snapshot, peers={'peer': exchange}))
    with TestClient(app) as client:
        assert client.post('/health', json={'safe': True, 'text': 'health',
                                           'downstream': 'peer'}).json()['success'] is True
        assert client.post('/health', json={'safe': True, 'text': 'bad',
                                           'downstream': 'peer'}).json()['success'] is False
    assert requests == [200, 200]
    class State(TypedDict):
        agenthealth: object
    report = object()
    class Client:
        async def async_check(self, configuration):
            return report
    builder = StateGraph(State)
    builder.add_node('health', health_node(Client(), 'operator.yaml'))
    builder.add_edge(START, 'health')
    builder.add_edge('health', END)
    assert asyncio.run(builder.compile().ainvoke({}))['agenthealth'] is report


def test_passive_requests_never_execute_graph_or_peer():
    class ForbiddenGraph:
        async def ainvoke(self, state):
            raise AssertionError('passive graph execution')
    async def forbidden_peer(text):
        raise AssertionError('passive peer execution')
    app = FastAPI()
    app.include_router(langgraph_router(ForbiddenGraph(), snapshot,
                                       peers={'peer': forbidden_peer}))
    with TestClient(app) as client:
        assert client.head('/health').status_code == 200
        assert client.get('/health').json()['version'] == 'v1'
        assert client.post('/health', json={'safe': False, 'text': 'health',
                                           'downstream': 'peer'}).status_code == 400


@pytest.mark.parametrize('peers', [{'': lambda _: None}, {' ': lambda _: None},
                                    {'peer': None}, {'x' * 1025: lambda _: None}])
def test_invalid_peer_configuration(peers):
    with pytest.raises(ValueError):
        langgraph_router(Graph(), snapshot, peers=peers)
