import {
    ActionRowBuilder,
    ButtonInteraction,
    MessageFlags,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
    type ModalActionRowComponentBuilder,
} from 'discord.js';
import { GameState, motusGames } from '../../../utils/games/motus';
import { errorEmbedGenerator } from '../../utils/embeds';

export async function handleMotusTry(interaction: ButtonInteraction): Promise<void> {
    const message = interaction.message;
    const game = motusGames.get(message.id);

    if (!game) {
        await interaction.reply({
            embeds: [errorEmbedGenerator('Partie de Motus terminée ou inexistante.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    if (game.state !== GameState.PLAYING) {
        await interaction.reply({
            embeds: [errorEmbedGenerator('La partie de Motus est déjà terminée.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const wordLength = game.wordLength;

    // const modal = new ModalBuilder().setCustomId('handleMotusTryModal--' + message.id).setTitle('Essai Motus');
    const modal = new ModalBuilder().setCustomId(`handleMotusTryModal`).setTitle('Essai Motus');

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
