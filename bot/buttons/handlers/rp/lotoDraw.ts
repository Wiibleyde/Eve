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

    // Tirer un ticket au hasard
    const randomIndex = Math.floor(Math.random() * game.tickets.length);
    const winningTicket = game.tickets[randomIndex];

    if (!winningTicket) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Erreur lors du tirage au sort.')],
        });
        return;
    }

    // Le numéro du ticket est basé sur sa position (index + 1)
    const winningTicketNumber = randomIndex + 1;

    // Mettre à jour le jeu avec le gagnant
    await prisma.lotoGames.update({
        where: { uuid: gameUuid },
        data: {
            isActive: false,
            winnerUuid: winningTicket.playerUuid,
        },
    });

    const winnerName = winningTicket.player?.name ?? 'Nom inconnu';
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
    const salesLines = Array.from(salesBySeller.entries())
        .map(([sellerId, stats]) => `<@${sellerId}> : ${stats.count} ticket(s) - **${stats.total}$**`)
        .join('\n');

    // Mettre à jour le message original
    if (interaction.message) {
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
            },
        });

        if (updatedGame) {
            const embed = generateLotoEmbed(updatedGame, updatedGame.tickets);
            embed.addFields({
                name: '🎉 Gagnant',
                value: `**${winnerName}** remporte **${totalPrize}$** !`,
            });
            embed.addFields({
                name: '📊 Ventes par vendeur',
                value: salesLines,
            });

            await interaction.message.edit({
                embeds: [embed],
                components: [], // Retirer tous les boutons
            });
        }
    }

    await interaction.editReply({
        embeds: [
            successEmbedGenerator(
                `🎉 Le gagnant est **${winnerName}** !\nIl remporte **${totalPrize}$** avec le ticket numéro **${winningTicketNumber}**.`
            ),
        ],
    });
}
