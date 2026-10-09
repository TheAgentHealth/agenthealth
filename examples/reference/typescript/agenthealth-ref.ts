import { Router, json, type ErrorRequestHandler } from 'express';

type Document = {
  name: string; live: boolean; ready: boolean;
  capabilities: string[]; dependencies: string[];
};
type Task = { safe: true; text: string; downstream?: string };
type Outcome = { completed: boolean; success: boolean; downstream?: string };

// Apply your application's authorization middleware before this router.
// Probes must honor AbortSignal; a timeout cannot forcibly stop arbitrary JS.
export function agentHealth(snapshot: () => Document,
  probe?: (task: Task, signal: AbortSignal) => Promise<Outcome>) {
  const router = Router();
  router.head('/health', (_req, res) => { res.sendStatus(200); });
  router.get('/health', (_req, res) => {
    res.set('Cache-Control', 'no-store').json({ ...snapshot(), version: 'v1' });
  });
  router.post('/health', json({ limit: '64kb' }), async (req, res) => {
    if (!probe) { res.sendStatus(405); return; }
    const task = req.body;
    if (!task || Array.isArray(task) || task.safe !== true ||
        typeof task.text !== 'string' || !task.text.trim() ||
        (task.downstream !== undefined && typeof task.downstream !== 'string') ||
        Object.keys(task).some(k => !['safe', 'text', 'downstream'].includes(k))) {
      res.sendStatus(400); return;
    }
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout> | undefined;
    try {
      const timeout = new Promise<never>((_resolve, reject) => {
        timer = setTimeout(() => { controller.abort(); reject(new Error('deadline')); }, 1000);
      });
      const outcome = await Promise.race([probe(task, controller.signal), timeout]);
      res.set('Cache-Control', 'no-store').json(outcome);
    } catch { res.sendStatus(503); }
    finally { clearTimeout(timer); }
  });
  // Avoid Express's default error logging and HTML stack traces for malformed input.
  const parseError: ErrorRequestHandler = (error, _req, res, _next) => {
    res.sendStatus(error?.type === 'entity.too.large' ? 413 : 400);
  };
  router.use(parseError);
  return router;
}
