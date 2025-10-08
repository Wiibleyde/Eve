import { basicEmbedGenerator } from '@bot/utils/embeds';
import type { LotoGames, LotoPlayers, LotoTickets } from '@prisma/client';

type LotoTicketWithPlayer = LotoTickets & { player: Pick<LotoPlayers, 'name'> | null };

export function generateLotoEmbed(game: LotoGames, tickets: LotoTicketWithPlayer[]) {
    const embed = basicEmbedGenerator();
    embed.setTitle(`🎟️ Loto: ${game.name} 🎟️`);
    const ticketsSold = tickets.length;
    const descriptionLines: string[] = [
        "⚠️ __**Attention, l'écriture des noms est sensible à la case !**__",
        `Nombre de tickets vendus: **${ticketsSold}**`,
    ];

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

    return embed;
}
