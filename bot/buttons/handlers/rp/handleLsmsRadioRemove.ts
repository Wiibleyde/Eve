import {
    ActionRowBuilder,
    MessageFlags,
    StringSelectMenuBuilder,
    StringSelectMenuOptionBuilder,
    type ButtonInteraction,
} from 'discord.js';
import {
    formatRadioModalId,
    getRadioCustomIds,
    lsmsErrorEmbedGenerator,
    parseRadiosFromEmbed,
} from '../../../../utils/rp/lsms';

export async function handleLsmsRadioRemove(interaction: ButtonInteraction): Promise<void> {
    if (!interaction.channelId || !interaction.message) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator("Impossible de déterminer le salon d'interaction.")],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const currentEmbed = interaction.message.embeds[0];
    const radios = parseRadiosFromEmbed(currentEmbed);

    if (radios.length === 0) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator("Aucune radio n'est configurée.")],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const customIds = getRadioCustomIds();
    const select = new StringSelectMenuBuilder()
        .setCustomId(formatRadioModalId(customIds.removeSelectPrefix, interaction.channelId, interaction.message.id))
        .setPlaceholder('Sélectionnez une radio à retirer')
        .setMinValues(1)
        .setMaxValues(1);

    radios.slice(0, 25).forEach((radio) => {
        select.addOptions(
            new StringSelectMenuOptionBuilder()
                .setLabel(radio.name)
                .setValue(radio.name)
                .setDescription(radio.frequency.slice(0, 100))
        );
    });

    await interaction.reply({
        content:
            radios.length > 25
                ? 'Sélectionnez la radio à retirer (seules les 25 premières radios sont affichées).'
                : 'Sélectionnez la radio à retirer.',
        components: [new ActionRowBuilder<StringSelectMenuBuilder>().addComponents(select)],
        flags: [MessageFlags.Ephemeral],
    });
}
