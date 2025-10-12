import { currentMusic } from '@bot/events/handlers/presenceUpdate';
import type { Request, Response } from 'express';

export function currentMusicHandler(req: Request, res: Response) {
    res.json({ currentlyPlaying: currentMusic });
}
