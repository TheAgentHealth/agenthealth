"""Connect a dedicated safe LangGraph graph to the FastAPI reference handler."""
from agenthealth_ref import router


def langgraph_router(graph, snapshot):
    # Use a separate graph with only non-destructive health nodes. Do not attach
    # a general production agent with arbitrary tool access to this handler.
    async def probe(task):
        if task.get('downstream'):
            # A real downstream exchange must be implemented by the application.
            return {'completed': True, 'success': False}
        state = await graph.ainvoke({'health_probe': task['text']})
        return {'completed': True, 'success': state.get('healthy') is True}
    return router(snapshot, probe)
