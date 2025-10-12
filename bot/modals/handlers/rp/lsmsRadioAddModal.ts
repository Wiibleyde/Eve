import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import {
    canAddRadio,
    decodeRadioModalId,
    lsmsErrorEmbedGenerator,
    lsmsSuccessEmbedGenerator,
    parseRadiosFromEmbed,
    updateRadioMessage,
} from '../../../../utils/rp/lsms';

export async function lsmsRadioAddModal(interaction: ModalSubmitInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    const { channelId, messageId } = decodeRadioModalId(interaction.customId);

    if (!channelId || !messageId) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Impossible de retrouver le message cible.')],
        });
        return;
    }

    const channel = await interaction.client.channels.fetch(channelId).catch(() => null);

    if (!channel || !channel.isTextBased()) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le salon de destination est introuvable ou non textuel.')],
        });
        return;
    }

    const message = await channel.messages.fetch(messageId).catch(() => null);

    if (!message) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le message du gestionnaire est introuvable.')],
        });
        return;
    }

    const radioName = interaction.fields.getTextInputValue('radioName').trim();
    const radioFrequency = interaction.fields.getTextInputValue('radioFrequency').trim();

    if (radioName.length === 0 || radioFrequency.length === 0) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le nom et la fréquence doivent être renseignés.')],
        });
        return;
    }

    const radios = parseRadiosFromEmbed(message.embeds[0]);

    if (!canAddRadio(radios)) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le nombre maximum de radios a été atteint.')],
        });
        return;
    }

    const alreadyExists = radios.some((radio) => radio.name.toLowerCase() === radioName.toLowerCase());

    if (alreadyExists) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Une radio avec ce nom existe déjà.')],
        });
        return;
    }

    radios.push({ name: radioName, frequency: radioFrequency });

    const { embed, components } = updateRadioMessage(radios);

    await message.edit({
        embeds: [embed],
        components,
    });

    await interaction.editReply({
        embeds: [lsmsSuccessEmbedGenerator(`La radio **${radioName}** a été ajoutée.`)],
    });
}
