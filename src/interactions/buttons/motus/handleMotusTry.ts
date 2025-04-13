import { errorEmbed } from '@/utils/embeds';
import { games, GameState } from '@/utils/games/motus';
import {
    ActionRowBuilder,
    ButtonInteraction,
    ModalActionRowComponentBuilder,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
} from 'discord.js';

export async function handleMotusTry(interaction: ButtonInteraction) {
    const message = interaction.message;
    const game = games.get(message.id);

    if (!game) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Partie de Motus terminée ou inexistante.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    if (game.state !== GameState.PLAYING) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('La partie de Motus est déjà terminée.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const wordLength = game.wordLength;

    const modal = new ModalBuilder().setCustomId('handleMotusTryModal--' + message.id).setTitle('Essai Motus');

    const textInput = new TextInputBuilder()
        .setLabel(`Entrez un mot de ${wordLength} lettres`)
        .setPlaceholder('Mot')
        .setCustomId('motusTry')
        .setStyle(TextInputStyle.Short)
        .setMaxLength(wordLength)
        .setMinLength(wordLength)
        .setRequired(true);

    const actionRow = new ActionRowBuilder<ModalActionRowComponentBuilder>().addComponents(textInput);
    modal.addComponents(actionRow);

    await interaction.showModal(modal);
}
