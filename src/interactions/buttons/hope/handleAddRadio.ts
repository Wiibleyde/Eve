import {
    ActionRowBuilder,
    ButtonInteraction,
    CacheType,
    ModalActionRowComponentBuilder,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
} from 'discord.js';

export async function handleAddRadio(interaction: ButtonInteraction<CacheType>): Promise<void> {
    const modal = new ModalBuilder()
        .setCustomId('addRadioModal--' + interaction.message.id)
        .setTitle('Ajouter une radio');

    const name = new TextInputBuilder()
        .setCustomId('name')
        .setPlaceholder('Nom')
        .setLabel('Nom')
        .setStyle(TextInputStyle.Short)
        .setRequired(true);

    const frequency = new TextInputBuilder()
        .setCustomId('frequency')
        .setPlaceholder('0.0')
        .setLabel('Fréquence')
        .setStyle(TextInputStyle.Short)
        .setRequired(true);

    const nameRow = new ActionRowBuilder<ModalActionRowComponentBuilder>().addComponents(name);
    const frequencyRow = new ActionRowBuilder<ModalActionRowComponentBuilder>().addComponents(frequency);
    modal.addComponents(nameRow, frequencyRow);

    await interaction.showModal(modal);
}
