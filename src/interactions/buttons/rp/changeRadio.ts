import { RadioFrequencies } from '@/utils/rp/radio';
import {
    ActionRowBuilder,
    ButtonInteraction,
    CacheType,
    ModalActionRowComponentBuilder,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
} from 'discord.js';

/**
 * Handles the interaction for changing the radio frequency.
 *
 * @param interaction - The interaction object representing the button interaction.
 * @returns A promise that resolves when the interaction is handled.
 *
 * @throws Will reply with an error message if the radio frequency does not exist.
 */
export async function changeRadio(interaction: ButtonInteraction<CacheType>): Promise<void> {
    const radioName = interaction.customId.split('--')[1];
    const message = interaction.message;
    const embed = message.embeds[0];
    const radioFrequencies: RadioFrequencies[] = [];
    if (embed.fields) {
        for (const field of embed.fields) {
            if (field.name.startsWith('Radio')) {
                const radioName = field.name.split('Radio ')[1];
                const frequency = field.value.split(' : ')[1];
                radioFrequencies.push({ name: radioName, frequency: frequency });
            }
        }
    }

    let radioIndex = -1;
    for (const frequency of radioFrequencies) {
        if (frequency.name === radioName) {
            radioIndex = radioFrequencies.indexOf(frequency);
            break;
        }
    }

    if (radioIndex === -1) {
        await interaction.reply({ content: "La radio demandée n'existe pas.", ephemeral: true });
        return;
    }

    if (radioFrequencies.length === 0) {
        await interaction.reply({ content: "Aucune fréquence n'a été trouvée.", ephemeral: true });
        return;
    }

    const modal = new ModalBuilder()
        .setCustomId('changeRadioModal--' + interaction.message.id + '--' + radioIndex)
        .setTitle('Changer la fréquence de la radio ' + radioName);

    const frequency = new TextInputBuilder()
        .setCustomId('frequency')
        .setPlaceholder('0.0')
        .setValue(radioFrequencies[radioIndex].frequency)
        .setLabel('Fréquence')
        .setStyle(TextInputStyle.Short)
        .setRequired(true);

    const actionRow = new ActionRowBuilder<ModalActionRowComponentBuilder>().addComponents(frequency);
    modal.addComponents(actionRow);

    await interaction.showModal(modal);
}
