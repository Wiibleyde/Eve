import { RadioFrequencies } from '@/utils/rp/radio';
import {
    ActionRowBuilder,
    ButtonInteraction,
    CacheType,
    EmbedBuilder,
    MessageFlags,
    StringSelectMenuBuilder,
    StringSelectMenuOptionBuilder,
} from 'discord.js';

export async function handleRemoveRadio(interaction: ButtonInteraction<CacheType>): Promise<void> {
    const message = interaction.message;
    const embed = message.embeds[0];
    const companyName = embed.title?.split(' du ')[1];
    if (!companyName) {
        await interaction.reply({ content: "Nom de l'entreprise introuvable", flags: [MessageFlags.Ephemeral] });
        return;
    }
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

    const responseEmbed = new EmbedBuilder()
        .setTitle('Supprimer une radio')
        .setDescription('Sélectionner une radio à supprimer')
        .setColor(0xff0000)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    const names = new StringSelectMenuBuilder()
        .setCustomId('removeRadio--' + interaction.message.id)
        .setPlaceholder('Sélectionner une radio à supprimer')
        .setMinValues(1);

    for (const radio of radioFrequencies) {
        names.addOptions(new StringSelectMenuOptionBuilder().setLabel(radio.name).setValue(radio.name));
    }

    names.setMaxValues(radioFrequencies.length);

    const actionRow = new ActionRowBuilder<StringSelectMenuBuilder>().addComponents(names);

    await interaction.reply({ embeds: [responseEmbed], components: [actionRow], flags: [MessageFlags.Ephemeral] });
}
