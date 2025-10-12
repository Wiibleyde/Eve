import {
    ActionRowBuilder,
    MessageFlags,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
    type ButtonInteraction,
} from 'discord.js';
import {
    decodeRadioButtonId,
    encodeRadioName,
    formatRadioModalId,
    getRadioCustomIds,
    lsmsErrorEmbedGenerator,
    parseRadiosFromEmbed,
} from '../../../../utils/rp/lsms';

export async function handleLsmsRadioEdit(interaction: ButtonInteraction): Promise<void> {
    if (!interaction.channelId || !interaction.message) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator("Impossible de déterminer le salon d'interaction.")],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const radioName = decodeRadioButtonId(interaction.customId);

    if (!radioName) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator('Impossible de déterminer la radio à modifier.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const currentEmbed = interaction.message.embeds[0];
    const radios = parseRadiosFromEmbed(currentEmbed);
    const radioData = radios.find((radio) => radio.name.toLowerCase() === radioName.toLowerCase());

    if (!radioData) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator("Cette radio n'existe plus.")],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const customIds = getRadioCustomIds();
    const encodedName = encodeRadioName(radioName);

    const modal = new ModalBuilder()
        .setCustomId(
            formatRadioModalId(customIds.editModalPrefix, interaction.channelId, interaction.message.id, encodedName)
        )
        .setTitle(`Modifier ${radioName}`);

    const nameInput = new TextInputBuilder()
        .setCustomId('radioUpdatedName')
        .setLabel('Nom de la radio')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(100)
        .setValue(radioData.name)
        .setMinLength(2);

    const frequencyInput = new TextInputBuilder()
        .setCustomId('radioUpdatedFrequency')
        .setLabel('Fréquence')
        .setStyle(TextInputStyle.Short)
        .setRequired(true)
        .setMaxLength(100)
        .setValue(radioData.frequency);

    modal.addComponents(
        new ActionRowBuilder<TextInputBuilder>().addComponents(nameInput),
        new ActionRowBuilder<TextInputBuilder>().addComponents(frequencyInput)
    );

    await interaction.showModal(modal);
}
