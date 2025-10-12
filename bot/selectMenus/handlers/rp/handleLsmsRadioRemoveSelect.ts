import { MessageFlags, type StringSelectMenuInteraction } from 'discord.js';
import {
    decodeRadioModalId,
    lsmsErrorEmbedGenerator,
    lsmsSuccessEmbedGenerator,
    parseRadiosFromEmbed,
    updateRadioMessage,
} from '../../../../utils/rp/lsms';

export async function handleLsmsRadioRemoveSelect(interaction: StringSelectMenuInteraction): Promise<void> {
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

    const radioName = interaction.values[0]?.trim();

    if (!radioName) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune radio sélectionnée.')],
        });
        return;
    }

    const radios = parseRadiosFromEmbed(message.embeds[0]);
    const filteredRadios = radios.filter((radio) => radio.name.toLowerCase() !== radioName.toLowerCase());

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
        embeds: [lsmsSuccessEmbedGenerator(`La radio **${radioName}** a été retirée.`)],
    });
}
