import type { Request, Response } from 'express';

export function baseHandler(req: Request, res: Response) {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
}

export function healthHandler(req: Request, res: Response) {
    res.json({ health: 'good', uptime: process.uptime() });
}

export function infoHandler(req: Request, res: Response) {
    res.json({ app: 'Eve - API', version: '0.1.0', author: 'Wiibleyde' });
}

export function pingHandler(req: Request, res: Response) {
    res.json({ message: 'pong', timestamp: new Date().toISOString() });
}
