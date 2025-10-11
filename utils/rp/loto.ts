import { basicEmbedGenerator } from '@bot/utils/embeds';
import type { LotoGames, LotoPlayers, LotoTickets, LotoPrizes } from '@prisma/client';

type LotoTicketWithPlayer = LotoTickets & { player: Pick<LotoPlayers, 'name'> | null };
type LotoPrizeWithWinner = LotoPrizes & { winnerPlayer: Pick<LotoPlayers, 'name'> | null };

function chunkLines(lines: string[], maxLength = 1024): string[] {
    const chunks: string[] = [];
    let current = '';

    for (const line of lines) {
        const candidate = current.length === 0 ? line : `${current}\n${line}`;
        if (candidate.length > maxLength) {
            if (current.length === 0) {
                chunks.push(line.slice(0, maxLength));
                current = line.slice(maxLength).trimStart();
                continue;
            }

            chunks.push(current);
            current = line;
            continue;
        }

        current = candidate;
    }

    if (current.length > 0) {
        chunks.push(current);
    }

    return chunks;
}

export function generateLotoEmbed(
    game: LotoGames & { prizes: LotoPrizeWithWinner[] },
    tickets: LotoTicketWithPlayer[]
) {
    const embed = basicEmbedGenerator();
    embed.setTitle(`🎟️ Loto: ${game.name} 🎟️`);
    const ticketsSold = tickets.length;
    const descriptionLines: string[] = [
        "⚠️ __**Attention, l'écriture des noms est sensible à la case !**__",
        `Nombre de tickets vendus: **${ticketsSold}**`,
    ];

    if (game.prizes.length > 0) {
        descriptionLines.push(`Nombre de gains: **${game.prizes.length}**`);
    }

    if (ticketsSold > 0) {
        descriptionLines.push(`Cagnotte actuelle: **${ticketsSold * game.ticketPrice}**$`);
    }

    const ticketsPerPlayer = new Map<string, { name: string; count: number }>();

    for (const ticket of tickets) {
        const playerId = ticket.playerUuid;
        const playerName = ticket.player?.name ?? 'Nom inconnu';

        const entry = ticketsPerPlayer.get(playerId);
        if (entry) {
            entry.count += 1;
        } else {
            ticketsPerPlayer.set(playerId, { name: playerName, count: 1 });
        }
    }

    const sortedPlayers = Array.from(ticketsPerPlayer.values()).sort((a, b) => {
        if (b.count !== a.count) return b.count - a.count;
        return a.name.localeCompare(b.name, 'fr', { sensitivity: 'case' });
    });

    // Limiter à 50 joueurs affichés
    const finalSortedPlayers = sortedPlayers.slice(0, 50);

    if (finalSortedPlayers.length > 0) {
        const playersLines = finalSortedPlayers.map(({ name, count }) => `- ${name} (${count})`);
        descriptionLines.push('', ...playersLines);
    } else {
        descriptionLines.push('', 'Aucun ticket vendu pour le moment.');
    }

    embed.setDescription(descriptionLines.join('\n'));

    embed.setFooter({ text: `Prix du ticket: ${game.ticketPrice}$` });

    if (game.prizes.length > 0) {
        const sortedPrizes = [...game.prizes].sort((a, b) => a.position - b.position);

        const prizeLines = sortedPrizes.map((prize, index) => {
            const rank = index + 1;
            if (prize.winnerPlayer) {
                const ticketInfo = prize.winningTicketNumber ? ` (ticket n°${prize.winningTicketNumber})` : '';
                return `- **#${rank}** ${prize.label} — 🎉 ${prize.winnerPlayer.name}${ticketInfo}`;
            }

            if (game.isActive) {
                return `- **#${rank}** ${prize.label}`;
            }

            return `- **#${rank}** ${prize.label} — Non attribué`;
        });

        const chunks = chunkLines(prizeLines.length > 0 ? prizeLines : ['Aucun gain configuré.']);
        chunks.forEach((chunk, idx) => {
            embed.addFields({ name: idx === 0 ? '🎁 Gains' : '🎁 Gains (suite)', value: chunk });
        });
    }

    return embed;
}
