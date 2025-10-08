import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import { errorEmbedGenerator, successEmbedGenerator } from '@bot/utils/embeds';
import { prisma } from '@utils/core/database';
import { generateLotoEmbed } from '@utils/rp/loto';

export async function removeTicketsModal(interaction: ModalSubmitInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    const gameUuid = interaction.customId.split('--')[1];

    if (!gameUuid) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Impossible de trouver le jeu.')],
        });
        return;
    }

    const playerName = interaction.fields.getTextInputValue('playerName').trim();
    const ticketCountStr = interaction.fields.getTextInputValue('ticketCount').trim();

    const ticketCount = parseInt(ticketCountStr, 10);

    if (isNaN(ticketCount) || ticketCount <= 0) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Le nombre de tickets doit être un nombre entier positif.')],
        });
        return;
    }

    if (playerName.length === 0) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Le nom du joueur ne peut pas être vide.')],
        });
        return;
    }

    const game = await prisma.lotoGames.findUnique({
        where: { uuid: gameUuid },
    });

    if (!game) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Jeu introuvable.')],
        });
        return;
    }

    if (!game.isActive) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Ce jeu est terminé.')],
        });
        return;
    }

    // Trouver le joueur
    const player = await prisma.lotoPlayers.findFirst({
        where: {
            gameUuid: gameUuid,
            name: playerName,
        },
        include: {
            tickets: true,
        },
    });

    if (!player) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator(`Le joueur **${playerName}** n'existe pas dans ce jeu.`)],
        });
        return;
    }

    if (player.tickets.length < ticketCount) {
        await interaction.editReply({
            embeds: [
                errorEmbedGenerator(
                    `**${playerName}** n'a que ${player.tickets.length} ticket(s), impossible d'en retirer ${ticketCount}.`
                ),
            ],
        });
        return;
    }

    // Supprimer les tickets (on prend les premiers trouvés)
    const ticketsToDelete = player.tickets.slice(0, ticketCount).map((t) => t.uuid);

    await prisma.lotoTickets.deleteMany({
        where: {
            uuid: {
                in: ticketsToDelete,
            },
        },
    });

    // Si le joueur n'a plus de tickets, on peut le supprimer
    const remainingTickets = player.tickets.length - ticketCount;
    if (remainingTickets === 0) {
        await prisma.lotoPlayers.delete({
            where: { uuid: player.uuid },
        });
    }

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
        },
    });

    if (updatedGame && interaction.message) {
        const embed = generateLotoEmbed(updatedGame, updatedGame.tickets);
        await interaction.message.edit({
            embeds: [embed],
        });
    }

    await interaction.editReply({
        embeds: [
            successEmbedGenerator(
                `${ticketCount} ticket(s) retiré(s) pour **${playerName}**.\nMontant remboursé: **${ticketCount * game.ticketPrice}$**`
            ),
        ],
    });
}
