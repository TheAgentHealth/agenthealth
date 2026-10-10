"""Reference FastAPI contract. Application authentication must wrap these routes."""
import asyncio
import json
from fastapi import APIRouter, HTTPException, Request, Response


# Allow JSON escaping/framing, while bounding decoded fields separately.
MAX_TASK_BODY_BYTES = 512 * 1024


def router(snapshot, probe=None):
    """snapshot returns current declarations; probe is async and cooperatively cancellable."""
    routes = APIRouter()

    @routes.head('/health')
    async def head():
        return Response(status_code=200)

    @routes.get('/health')
    async def health():
        return {**snapshot(), 'version': 'v1'}

    @routes.post('/health')
    async def task(request: Request):
        if probe is None:
            raise HTTPException(405)
        body = bytearray()
        async for chunk in request.stream():
            if len(body) + len(chunk) > MAX_TASK_BODY_BYTES:
                raise HTTPException(413)
            body.extend(chunk)
        try:
            data = json.loads(body)
        except (ValueError, UnicodeDecodeError):
            raise HTTPException(400)
        if (not isinstance(data, dict) or data.get('safe') is not True
                or not isinstance(data.get('text'), str) or not data['text'].strip()
                or not set(data) <= {'safe', 'text', 'downstream'}
                or not isinstance(data.get('downstream', ''), str)):
            raise HTTPException(400)
        try:
            text_bytes = len(data['text'].encode('utf-8'))
            downstream = data.get('downstream', '')
            downstream.encode('utf-8')
        except UnicodeError:
            raise HTTPException(400) from None
        if (text_bytes > 65536 or len(downstream) > 1024
                or (downstream and not downstream.strip())):
            raise HTTPException(400)
        try:
            return await asyncio.wait_for(probe(data), timeout=1)
        except Exception:
            # Do not leak task text, exception messages or backend credentials.
            raise HTTPException(503) from None

    return routes
