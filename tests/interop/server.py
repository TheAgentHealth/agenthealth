"""Loopback-only official SDK servers used by interoperability CI."""
import argparse
import importlib.metadata
from uuid import uuid4


def mcp_server(args):
    from mcp.types import ToolAnnotations
    modern = importlib.metadata.version("mcp").split(".")[0] == "2"
    if modern:
        from mcp.server import MCPServer
        server = MCPServer("agenthealth-interop")
    else:
        from mcp.server.fastmcp import FastMCP
        server = FastMCP("agenthealth-interop", host="127.0.0.1", port=args.port)

    @server.tool(annotations=ToolAnnotations(readOnlyHint=True, destructiveHint=False))
    def health() -> str:
        """Read-only deterministic health response."""
        return "OK"

    if args.transport == "stdio":
        server.run(transport="stdio")
    elif modern:
        server.run(transport="streamable-http", host="127.0.0.1", port=args.port,
                   json_response=True)
    else:
        server.run(transport="streamable-http")


def a2a_server(args):
    import uvicorn
    from a2a.server.agent_execution import AgentExecutor
    from a2a.server.tasks import InMemoryTaskStore
    from a2a.types import AgentCard, AgentCapabilities, AgentSkill, Message, Part
    modern = importlib.metadata.version("a2a-sdk").split(".")[0] == "1"
    endpoint = f"http://127.0.0.1:{args.port}/rpc"
    skill = AgentSkill(id="health", name="Health", description="Read-only health", tags=["health"])
    if modern:
        from a2a.types import AgentInterface, Role
        from a2a.server.request_handlers import DefaultRequestHandlerV2
        from a2a.server.routes.agent_card_routes import create_agent_card_routes
        from a2a.server.routes.jsonrpc_routes import create_jsonrpc_routes
        from starlette.applications import Starlette
        card = AgentCard(name="health", description="Health server", version="1",
                         supported_interfaces=[AgentInterface(url=endpoint, protocol_binding="JSONRPC", protocol_version="1.0")],
                         capabilities=AgentCapabilities(), skills=[skill],
                         default_input_modes=["text/plain"], default_output_modes=["text/plain"])
    else:
        from a2a.types import TextPart
        from a2a.server.apps import A2AStarletteApplication
        from a2a.server.request_handlers import DefaultRequestHandler
        card = AgentCard(name="health", description="Health server", version="1", url=endpoint,
                         protocol_version="0.3.0", capabilities=AgentCapabilities(), skills=[skill],
                         default_input_modes=["text/plain"], default_output_modes=["text/plain"])

    class HealthExecutor(AgentExecutor):
        async def execute(self, context, event_queue):
            if modern:
                message = Message(message_id=str(uuid4()), context_id=context.context_id,
                                  role=Role.ROLE_AGENT, parts=[Part(text="OK")])
            else:
                message = Message(message_id=str(uuid4()), role="agent", parts=[Part(root=TextPart(text="OK"))])
            await event_queue.enqueue_event(message)

        async def cancel(self, context, event_queue):
            raise NotImplementedError("health fixture does not create tasks")

    if modern:
        handler = DefaultRequestHandlerV2(HealthExecutor(), InMemoryTaskStore(), card)
        app = Starlette(routes=create_agent_card_routes(card) + create_jsonrpc_routes(handler, "/rpc"))
    else:
        handler = DefaultRequestHandler(HealthExecutor(), InMemoryTaskStore())
        app = A2AStarletteApplication(card, handler).build(rpc_url="/rpc")
    uvicorn.run(app, host="127.0.0.1", port=args.port, log_level="warning")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("protocol", choices=["a2a", "mcp"])
    parser.add_argument("--port", type=int, default=8000)
    parser.add_argument("--transport", choices=["stdio", "http"], default="http")
    args = parser.parse_args()
    (a2a_server if args.protocol == "a2a" else mcp_server)(args)
