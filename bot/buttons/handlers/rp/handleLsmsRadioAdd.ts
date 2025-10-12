import {
    ActionRowBuilder,
    MessageFlags,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
    type ButtonInteraction,
} from 'discord.js';
import {
    canAddRadio,
    formatRadioModalId,
    getRadioCustomIds,
    lsmsErrorEmbedGenerator,
    parseRadiosFromEmbed,
} from '../../../../utils/rp/lsms';

export async function handleLsmsRadioAdd(interaction: ButtonInteraction): Promise<void> {
    if (!interaction.channelId || !interaction.message) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator("Impossible de déterminer le salon d'interaction.")],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const currentEmbed = interaction.message.embeds[0];
    const radios = parseRadiosFromEmbed(currentEmbed);

    if (!canAddRadio(radios)) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator('Le nombre maximum de radios a été atteint.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const customIds = getRadioCustomIds();
    const modal = new ModalBuilder()
        .setCustomId(formatRadioModalId(customIds.addModalPrefix, interaction.channelId, interaction.message.id))
        .setTitle('Ajouter une radio');

    const nameInput = new TextInputBuilder()
        .setCustomId('radioName')
        .setLabel('Nom de la radio')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(100)
        .setPlaceholder('Ex: SAMU')
        .setMinLength(2);

    const frequencyInput = new TextInputBuilder()
        .setCustomId('radioFrequency')
        .setLabel('Fréquence')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(100)
        .setPlaceholder('Ex: 155.50');

    modal.addComponents(
        new ActionRowBuilder<TextInputBuilder>().addComponents(nameInput),
        new ActionRowBuilder<TextInputBuilder>().addComponents(frequencyInput)
    );

    await interaction.showModal(modal);
}
