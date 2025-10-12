import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import {
    decodeRadioModalId,
    decodeRadioName,
    lsmsErrorEmbedGenerator,
    lsmsSuccessEmbedGenerator,
    parseRadiosFromEmbed,
    updateRadioMessage,
} from '../../../../utils/rp/lsms';

export async function lsmsRadioEditModal(interaction: ModalSubmitInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    const { channelId, messageId, extra } = decodeRadioModalId(interaction.customId);

    if (!channelId || !messageId || !extra) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Impossible de retrouver le message cible.')],
        });
        return;
    }

    let originalName: string;

    try {
        originalName = decodeRadioName(extra);
    } catch {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Impossible de décoder les informations de la radio.')],
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

    const updatedName = interaction.fields.getTextInputValue('radioUpdatedName').trim();
    const updatedFrequency = interaction.fields.getTextInputValue('radioUpdatedFrequency').trim();

    if (updatedName.length === 0 || updatedFrequency.length === 0) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le nom et la fréquence doivent être renseignés.')],
        });
        return;
    }

    const radios = parseRadiosFromEmbed(message.embeds[0]);
    const radioIndex = radios.findIndex((radio) => radio.name.toLowerCase() === originalName.toLowerCase());

    if (radioIndex === -1) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator("Cette radio n'existe plus.")],
        });
        return;
    }

    const nameConflict = radios.some(
        (radio, index) => index !== radioIndex && radio.name.toLowerCase() === updatedName.toLowerCase()
    );

    if (nameConflict) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Une radio avec ce nom existe déjà.')],
        });
        return;
    }

    radios[radioIndex] = { name: updatedName, frequency: updatedFrequency };

    const { embed, components } = updateRadioMessage(radios);

    await message.edit({
        embeds: [embed],
        components,
    });

    await interaction.editReply({
        embeds: [lsmsSuccessEmbedGenerator(`La radio **${updatedName}** a été mise à jour.`)],
    });
}
