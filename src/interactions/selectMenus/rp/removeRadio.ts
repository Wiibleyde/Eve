import { creatEmbedForRadio, RadioFrequencies } from '@/utils/rp/radio';
import { StringSelectMenuInteraction, TextChannel } from 'discord.js';

export async function removeRadio(interaction: StringSelectMenuInteraction): Promise<void> {
    const selectedValues = interaction.values;
    const messageId = interaction.customId.split('--')[1];

    const channel = interaction.channel as TextChannel;
    const message = await channel.messages.fetch(messageId);
    if (!message) {
        await interaction.reply({ content: 'Message introuvable', ephemeral: true });
        return;
    }

    const embed = message.embeds[0];
    const companyName = embed.title?.split(' du ')[1];
    if (!companyName) {
        await interaction.reply({ content: "Nom de l'entreprise introuvable", ephemeral: true });
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

    const newRadioFrequencies: RadioFrequencies[] = [];
    for (const radio of radioFrequencies) {
        if (!selectedValues.includes(radio.name)) {
            newRadioFrequencies.push(radio);
        }
    }

    const { embed: newEmbed, actionRow, files } = creatEmbedForRadio(interaction, companyName, newRadioFrequencies);

    await message.edit({ embeds: [newEmbed], components: [actionRow], files: files });

    await interaction.reply({ content: 'Radio(s) supprimée(s)', ephemeral: true });
}
