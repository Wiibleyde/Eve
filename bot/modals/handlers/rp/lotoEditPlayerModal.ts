import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import { errorEmbedGenerator, successEmbedGenerator } from '@bot/utils/embeds';
import { prisma } from '@utils/core/database';
import { generateLotoEmbed } from '@utils/rp/loto';

export async function lotoEditPlayerModal(interaction: ModalSubmitInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    const gameUuid = interaction.customId.split('--')[1];

    if (!gameUuid) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Impossible de trouver le jeu.')],
        });
        return;
    }

    const oldName = interaction.fields.getTextInputValue('oldName').trim();
    const newName = interaction.fields.getTextInputValue('newName').trim();

    if (oldName.length === 0 || newName.length === 0) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator('Les noms ne peuvent pas être vides.')],
        });
        return;
    }

    if (oldName === newName) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator("Le nouveau nom doit être différent de l'ancien.")],
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

    // Trouver le joueur avec l'ancien nom
    const player = await prisma.lotoPlayers.findFirst({
        where: {
            gameUuid: gameUuid,
            name: oldName,
        },
    });

    if (!player) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator(`Le joueur **${oldName}** n'existe pas dans ce jeu.`)],
        });
        return;
    }

    // Vérifier qu'un joueur avec le nouveau nom n'existe pas déjà
    const existingPlayer = await prisma.lotoPlayers.findFirst({
        where: {
            gameUuid: gameUuid,
            name: newName,
        },
    });

    if (existingPlayer) {
        await interaction.editReply({
            embeds: [errorEmbedGenerator(`Un joueur nommé **${newName}** existe déjà dans ce jeu.`)],
        });
        return;
    }

    // Mettre à jour le nom
    await prisma.lotoPlayers.update({
        where: { uuid: player.uuid },
        data: { name: newName },
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
        embeds: [successEmbedGenerator(`Le nom **${oldName}** a été changé en **${newName}**.`)],
    });
}
