"""LangGraph health resource and consumer node; no framework imports at SDK startup."""
from collections.abc import Mapping


def langgraph_router(graph, snapshot, *, peers=None):
    """Expose a dedicated safe graph and explicitly allowlisted async peer probes.

    Each peer callback receives health text and must perform the real authorized
    exchange, returning exactly True only on success. No URL discovery occurs.
    The FastAPI handler applies a one-second cooperative deadline to all work.
    """
    from .fastapi import router

    selected_peers = dict(peers or {})
    if any(not isinstance(key, str) or not key.strip() or len(key) > 1024
           or not callable(value) for key, value in selected_peers.items()):
        raise ValueError("peers must map nonempty selectors to async callbacks")

    async def probe(task):
        downstream = task.get('downstream')
        if downstream:
            callback = selected_peers.get(downstream)
            if callback is None:
                return {'completed': True, 'success': False}
            success = await callback(task['text'])
            return {'completed': True, 'success': success is True,
                    'downstream': downstream}
        state = await graph.ainvoke({'health_probe': task['text']})
        return {'completed': True,
                'success': isinstance(state, Mapping) and state.get('healthy') is True}

    return router(snapshot, probe)


def health_node(client, configuration, *, output_key='agenthealth'):
    """Create an async StateGraph node returning the complete typed CheckReport.

    Configuration is operator-owned, never read from graph state. Health exit
    codes remain reports; SDK/tool failures and cancellation propagate normally.
    """
    if not isinstance(output_key, str) or not output_key.strip():
        raise ValueError('output_key must be nonempty')

    async def node(_state):
        report = await client.async_check(configuration)
        return {output_key: report}

    return node
