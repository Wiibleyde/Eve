import type { Request, Response } from 'express';
import { prisma } from '@utils/core/database';

export async function lotoStatsHandler(req: Request, res: Response) {
    try {
        const { lotoUuid } = req.query;

        if (lotoUuid && typeof lotoUuid !== 'string') {
            return res.status(400).json({ error: 'Invalid lotoUuid parameter' });
        }

        const whereClause = lotoUuid ? { uuid: lotoUuid } : {};

        const totalGames = await prisma.lotoGames.count({ where: whereClause });
        const totalTickets = await prisma.lotoTickets.count({
            where: lotoUuid ? { gameUuid: lotoUuid } : {},
        });

        res.json({ totalGames, totalTickets });
    } catch (error) {
        console.error('Error fetching loto stats:', error);
        res.status(500).json({ error: 'Internal Server Error' });
    }
}

export async function lotoWinnersHandler(req: Request, res: Response) {
    try {
        const { lotoUuid } = req.query;

        if (lotoUuid && typeof lotoUuid !== 'string') {
            return res.status(400).json({ error: 'Invalid lotoUuid parameter' });
        }

        const whereClause = lotoUuid ? { winnerUuid: { not: null }, uuid: lotoUuid } : { winnerUuid: { not: null } };

        const winners = await prisma.lotoGames.findMany({
            where: whereClause,
            include: {
                winner: {
                    select: {
                        name: true,
                    },
                },
            },
        });

        const formattedWinners = winners.map((game) => ({
            gameUuid: game.uuid,
            winnerName: game.winner ? game.winner.name : 'Inconnu',
        }));

        res.json({ winners: formattedWinners });
    } catch (error) {
        console.error('Error fetching loto winners:', error);
        res.status(500).json({ error: 'Internal Server Error' });
    }
}
