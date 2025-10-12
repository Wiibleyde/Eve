import { MessageFlags, type ModalSubmitInteraction } from 'discord.js';
import {
    decodeRadioModalId,
    lsmsErrorEmbedGenerator,
    lsmsSuccessEmbedGenerator,
    parseRadiosFromEmbed,
    updateRadioMessage,
} from '../../../../utils/rp/lsms';

export async function lsmsRadioRemoveModal(interaction: ModalSubmitInteraction): Promise<void> {
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

    const radioNameToRemove = interaction.fields.getTextInputValue('radioNameToRemove').trim();

    if (radioNameToRemove.length === 0) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Le nom de la radio doit être renseigné.')],
        });
        return;
    }

    const radios = parseRadiosFromEmbed(message.embeds[0]);

    const filteredRadios = radios.filter((radio) => radio.name.toLowerCase() !== radioNameToRemove.toLowerCase());

    if (filteredRadios.length === radios.length) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune radio ne correspond à ce nom.')],
        });
        return;
    }

    const { embed, components } = updateRadioMessage(filteredRadios);

    await message.edit({
        embeds: [embed],
        components,
    });

    await interaction.editReply({
        embeds: [lsmsSuccessEmbedGenerator(`La radio **${radioNameToRemove}** a été retirée.`)],
    });
}
