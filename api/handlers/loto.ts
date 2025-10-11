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
        const totalPrizes = await prisma.lotoPrizes.count({
            where: lotoUuid ? { gameUuid: lotoUuid } : {},
        });

        res.json({ totalGames, totalTickets, totalPrizes });
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

        const whereClause = lotoUuid ? { gameUuid: lotoUuid } : {};

        const prizes = await prisma.lotoPrizes.findMany({
            where: {
                winnerPlayerUuid: { not: null },
                ...whereClause,
            },
            include: {
                game: {
                    select: {
                        uuid: true,
                        name: true,
                    },
                },
                winnerPlayer: {
                    select: {
                        name: true,
                    },
                },
            },
            orderBy: [{ drawnAt: 'desc' }, { position: 'asc' }],
        });

        const formattedWinners = prizes.map((prize) => ({
            gameUuid: prize.game.uuid,
            gameName: prize.game.name,
            prizeLabel: prize.label,
            winnerName: prize.winnerPlayer ? prize.winnerPlayer.name : 'Inconnu',
            winningTicketNumber: prize.winningTicketNumber,
        }));

        res.json({ winners: formattedWinners });
    } catch (error) {
        console.error('Error fetching loto winners:', error);
        res.status(500).json({ error: 'Internal Server Error' });
    }
}
