import { MessageFlags, type ButtonInteraction } from 'discord.js';
import { errorEmbedGenerator, successEmbedGenerator } from '@bot/utils/embeds';
import { prisma } from '@utils/core/database';
import { hasPermission } from '@utils/permission';
import { PermissionFlagsBits } from 'discord.js';
import { generateLotoEmbed } from '@utils/rp/loto';

export async function lotoDraw(interaction: ButtonInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    if (!(await hasPermission(interaction, [PermissionFlagsBits.PinMessages]))) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator("Vous n'avez pas la permission d'effectuer cette action.")],
        });
        return;
    }

    const gameUuid = interaction.customId.split('--')[1];

    if (!gameUuid) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Impossible de trouver le jeu.')],
        });
        return;
    }

    const game = await prisma.lotoGames.findUnique({
        where: { uuid: gameUuid },
        include: {
            tickets: {
                include: {
                    player: {
                        select: {
                            name: true,
                        },
                    },
                },
            },
            prizes: {
                orderBy: { position: 'asc' },
                include: {
                    winnerPlayer: {
                        select: {
                            name: true,
                        },
                    },
                },
            },
        },
    });

    if (!game) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Jeu introuvable.')],
        });
        return;
    }

    if (!game.isActive) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Ce jeu est déjà terminé.')],
        });
        return;
    }

    if (game.tickets.length === 0) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Aucun ticket vendu, impossible de tirer un gagnant.')],
        });
        return;
    }

    if (game.prizes.length === 0) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Aucun gain configuré pour ce loto.')],
        });
        return;
    }

    // Count unique players
    const uniquePlayers = new Set(game.tickets.map((ticket) => ticket.playerUuid));
    const uniquePlayerCount = uniquePlayers.size;

    if (game.prizes.length > uniquePlayerCount) {
        await interaction.editReply({
            embeds: [
                errorEmbedGenerator(
                    `Il faut au moins autant de joueurs uniques que de gains (joueurs: ${uniquePlayerCount}, gains: ${game.prizes.length}).`
                ),
            ],
        });
        return;
    }

    const sortedPrizes = [...game.prizes].sort((a, b) => a.position - b.position);
    const availableTickets = [...game.tickets];
    const now = new Date();
    const winningPlayerUuids = new Set<string>();

    const assignments = sortedPrizes.map((prize) => {
        // Filter out tickets from players who have already won
        const eligibleTickets = availableTickets.filter((ticket) => !winningPlayerUuids.has(ticket.playerUuid));

        if (eligibleTickets.length === 0) {
            throw new Error('Pas assez de joueurs uniques pour tous les gains.');
        }

        const randomIndex = Math.floor(Math.random() * eligibleTickets.length);
        const winningTicket = eligibleTickets[randomIndex];

        if (!winningTicket) {
            throw new Error('Erreur lors du tirage au sort des gains.');
        }

        // Mark this player as having won
        winningPlayerUuids.add(winningTicket.playerUuid);

        // Remove the specific winning ticket from availableTickets
        const ticketIndex = availableTickets.findIndex((ticket) => ticket.uuid === winningTicket.uuid);
        if (ticketIndex !== -1) {
            availableTickets.splice(ticketIndex, 1);
        }

        const ticketNumber = game.tickets.findIndex((ticket) => ticket.uuid === winningTicket.uuid) + 1;

        return {
            prizeUuid: prize.uuid,
            prizeLabel: prize.label,
            ticket: winningTicket,
            ticketNumber,
        };
    });

    await prisma.$transaction([
        prisma.lotoGames.update({
            where: { uuid: gameUuid },
            data: {
                isActive: false,
            },
        }),
        ...assignments.map((assignment) =>
            prisma.lotoPrizes.update({
                where: { uuid: assignment.prizeUuid },
                data: {
                    winnerPlayerUuid: assignment.ticket.playerUuid,
                    winningTicketNumber: assignment.ticketNumber,
                    drawnAt: now,
                },
            })
        ),
    ]);

    const totalPrize = game.tickets.length * game.ticketPrice;

    // Calculer les statistiques de vente par vendeur
    const salesBySeller = new Map<string, { count: number; total: number }>();
    for (const ticket of game.tickets) {
        const existing = salesBySeller.get(ticket.sellerId);
        if (existing) {
            existing.count += 1;
            existing.total += game.ticketPrice;
        } else {
            salesBySeller.set(ticket.sellerId, { count: 1, total: game.ticketPrice });
        }
    }

    // Créer le texte des ventes par vendeur
    const sellerSummaries = Array.from(salesBySeller.entries()).map(
        ([sellerId, stats]) => `<@${sellerId}> : ${stats.count} ticket(s) - **${stats.total}$**`
    );
    const salesLines = sellerSummaries.join('\n');
    const salesFieldValue = salesLines.length > 0 ? salesLines : 'Aucune vente enregistrée.';

    // Mettre à jour le message original
    const updatedGame = await prisma.lotoGames.findUnique({
        where: { uuid: gameUuid },
        include: {
            tickets: {
                include: {
                    player: {
                        select: {
                            name: true,
                        },
                    },
                },
            },
            prizes: {
                orderBy: { position: 'asc' },
                include: {
                    winnerPlayer: {
                        select: {
                            name: true,
                        },
                    },
                },
            },
        },
    });

    if (updatedGame && interaction.message) {
        const embed = generateLotoEmbed(updatedGame, updatedGame.tickets);
        embed.addFields({
            name: '📊 Ventes par vendeur',
            value: salesFieldValue,
        });

        await interaction.message.edit({
            embeds: [embed],
            components: [],
        });
    }

    const winnerLines = (
        updatedGame?.prizes.map((prize, index) => {
            const rank = index + 1;
            if (prize.winnerPlayer) {
                const ticketInfo = prize.winningTicketNumber ? ` (ticket n°${prize.winningTicketNumber})` : '';
                return `#${rank} ${prize.label} → **${prize.winnerPlayer.name}**${ticketInfo}`;
            }

            return `#${rank} ${prize.label} → Non attribué`;
        }) ??
        assignments.map((assignment, index) => {
            const playerName = assignment.ticket.player?.name ?? 'Nom inconnu';
            return `#${index + 1} ${assignment.prizeLabel} → **${playerName}** (ticket n°${assignment.ticketNumber})`;
        })
    ).join('\n');

    await interaction.editReply({
        embeds: [successEmbedGenerator(`🎉 Tirage effectué !\n${winnerLines}\n\nCagnotte totale: **${totalPrize}$**.`)],
    });
}
