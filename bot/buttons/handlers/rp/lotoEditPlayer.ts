import {
    ActionRowBuilder,
    MessageFlags,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
    type ButtonInteraction,
} from 'discord.js';
import { errorEmbedGenerator } from '@bot/utils/embeds';
import { prisma } from '@utils/core/database';

export async function lotoEditPlayer(interaction: ButtonInteraction): Promise<void> {
    const gameUuid = interaction.customId.split('--')[1];

    if (!gameUuid) {
        await interaction.reply({
            embeds: [errorEmbedGenerator('Impossible de trouver le jeu.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const game = await prisma.lotoGames.findUnique({
        where: { uuid: gameUuid },
    });

    if (!game) {
        await interaction.reply({
            embeds: [errorEmbedGenerator('Jeu introuvable.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    if (!game.isActive) {
        await interaction.reply({
            embeds: [errorEmbedGenerator('Ce jeu est terminé.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const modal = new ModalBuilder()
        .setCustomId(`lotoEditPlayerModal--${gameUuid}`)
        .setTitle("Modifier le nom d'un participant");

    const oldNameInput = new TextInputBuilder()
        .setCustomId('oldName')
        .setLabel('Nom actuel du joueur')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(50)
        .setPlaceholder('Ex: John Doe');

    const newNameInput = new TextInputBuilder()
        .setCustomId('newName')
        .setLabel('Nouveau nom')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(50)
        .setPlaceholder('Ex: Jane Doe');

    const firstActionRow = new ActionRowBuilder<TextInputBuilder>().addComponents(oldNameInput);
    const secondActionRow = new ActionRowBuilder<TextInputBuilder>().addComponents(newNameInput);

    modal.addComponents(firstActionRow, secondActionRow);

    await interaction.showModal(modal);
}
