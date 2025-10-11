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

export async function removeTickets(interaction: ButtonInteraction): Promise<void> {
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

    const modal = new ModalBuilder().setCustomId(`removeTicketsModal--${gameUuid}`).setTitle('Retirer des tickets');

    const playerNameInput = new TextInputBuilder()
        .setCustomId('playerName')
        .setLabel('Nom du joueur')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(50)
        .setPlaceholder('Ex: John Doe');

    const ticketCountInput = new TextInputBuilder()
        .setCustomId('ticketCount')
        .setLabel('Nombre de tickets à retirer')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(5)
        .setPlaceholder('Ex: 2');

    const firstActionRow = new ActionRowBuilder<TextInputBuilder>().addComponents(playerNameInput);
    const secondActionRow = new ActionRowBuilder<TextInputBuilder>().addComponents(ticketCountInput);

    modal.addComponents(firstActionRow, secondActionRow);

    await interaction.showModal(modal);
}
