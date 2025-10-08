import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import { errorEmbedGenerator, successEmbedGenerator } from '@bot/utils/embeds';
import { prisma } from '@utils/core/database';
import { generateLotoEmbed } from '@utils/rp/loto';

export async function lotoBuyModal(interaction: ModalSubmitInteraction): Promise<void> {
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
    const sellerId = interaction.user.id;

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

    // Créer ou trouver le joueur
    let player = await prisma.lotoPlayers.findFirst({
        where: {
            gameUuid: gameUuid,
            name: playerName,
        },
    });

    if (!player) {
        player = await prisma.lotoPlayers.create({
            data: {
                gameUuid: gameUuid,
                name: playerName,
            },
        });
    } else {
        // Mettre à jour le lastPlay
        await prisma.lotoPlayers.update({
            where: { uuid: player.uuid },
            data: { lastPlay: new Date() },
        });
    }

    // Créer les tickets
    const ticketsToCreate = [];
    for (let i = 0; i < ticketCount; i++) {
        ticketsToCreate.push({
            playerUuid: player.uuid,
            gameUuid: gameUuid,
            sellerId: sellerId,
        });
    }

    await prisma.lotoTickets.createMany({
        data: ticketsToCreate,
    });

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
                `${ticketCount} ticket(s) ajouté(s) pour **${playerName}** par <@${sellerId}>.\nMontant total: **${ticketCount * game.ticketPrice}$**`
            ),
        ],
    });
}
