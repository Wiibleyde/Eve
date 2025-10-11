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

    const now = new Date();
    const cooldownMinutes = 'cooldownMinutes' in game ? (game.cooldownMinutes ?? 0) : 0;

    // Créer ou trouver le joueur
    let player = await prisma.lotoPlayers.findFirst({
        where: {
            gameUuid: gameUuid,
            name: playerName,
        },
    });

    if (player && cooldownMinutes > 0) {
        const nextAllowedTime = new Date(player.lastPlay.getTime() + cooldownMinutes * 60 * 1000);

        if (nextAllowedTime > now) {
            const remainingMs = nextAllowedTime.getTime() - now.getTime();
            const totalRemainingMinutes = Math.ceil(remainingMs / 60000);
            const hours = Math.floor(totalRemainingMinutes / 60);
            const minutes = totalRemainingMinutes % 60;

            const remainingParts: string[] = [];
            if (hours > 0) {
                remainingParts.push(`${hours} heure${hours > 1 ? 's' : ''}`);
            }
            if (minutes > 0) {
                remainingParts.push(`${minutes} minute${minutes > 1 ? 's' : ''}`);
            }

            const remainingText = remainingParts.length > 0 ? remainingParts.join(' et ') : "moins d'une minute";

            await interaction.editReply({
                embeds: [
                    errorEmbedGenerator(
                        `**${playerName}** doit patienter encore ${remainingText} avant de pouvoir racheter des tickets.`
                    ),
                ],
            });
            return;
        }
    }

    if (!player) {
        player = await prisma.lotoPlayers.create({
            data: {
                gameUuid: gameUuid,
                name: playerName,
                lastPlay: now,
            },
        });
    } else {
        // Mettre à jour le lastPlay
        player = await prisma.lotoPlayers.update({
            where: { uuid: player.uuid },
            data: { lastPlay: now },
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
