import { errorEmbed, successEmbed } from '@/utils/embeds';
import { creatEmbedForRadio, RadioFrequencies } from '@/utils/rp/radio';
import { MessageFlags, ModalSubmitInteraction, TextChannel } from 'discord.js';

export async function addRadioModal(interaction: ModalSubmitInteraction): Promise<void> {
    const name = interaction.fields.getTextInputValue('name');
    const frequency = interaction.fields.getTextInputValue('frequency');
    const newFrequency: RadioFrequencies = { name: name, frequency: frequency };
    const messageId = interaction.customId.split('--')[1];

    const channel = interaction.channel as TextChannel;
    const message = await channel.messages.fetch(messageId);
    if (!message) {
        await interaction.reply({ content: 'Message introuvable', flags: [MessageFlags.Ephemeral] });
        return;
    }

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

    for (const radio of radioFrequencies) {
        if (radio.name === name) {
            await interaction.reply({
                embeds: [errorEmbed(interaction, new Error('Nom déjà utilisé'))],
                ephemeral: true,
            });
            return;
        }
    }

    radioFrequencies.push(newFrequency);

    const { embed: newEmbed, actionRow, files } = creatEmbedForRadio(interaction, companyName, radioFrequencies);

    await message.edit({ embeds: [newEmbed], components: [actionRow], files: files });

    await interaction.reply({ embeds: [successEmbed(interaction, 'Radio ajoutée')], flags: [MessageFlags.Ephemeral] });
}
