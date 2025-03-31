import { successEmbed } from '@/utils/embeds';
import { creatEmbedForRadio, RadioFrequencies } from '@/utils/rp/radio';
import { MessageFlags, ModalSubmitInteraction, TextChannel } from 'discord.js';

/**
 * Handles the submission of a modal to change the radio frequency.
 *
 * @param interaction - The interaction object representing the modal submission.
 * @returns A promise that resolves when the radio frequency has been successfully changed and the user has been notified.
 */
export async function changeRadioModal(interaction: ModalSubmitInteraction) {
    const frequency = interaction.fields.getTextInputValue('frequency');
    const messageId = interaction.customId.split('--')[1];
    const radioIndex = parseInt(interaction.customId.split('--')[2]);

    // Fetch message from database
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

    if (radioIndex === -1) {
        await interaction.reply({ content: "La radio demandée n'existe pas.", flags: [MessageFlags.Ephemeral] });
        return;
    }

    if (radioFrequencies.length === 0) {
        await interaction.reply({ content: "Aucune fréquence n'a été trouvée.", flags: [MessageFlags.Ephemeral] });
        return;
    }

    const radio = radioFrequencies[radioIndex];
    radio.frequency = frequency;

    const { embed: newEmbed, actionRow, files } = creatEmbedForRadio(interaction, companyName, radioFrequencies);

    await message.edit({ embeds: [newEmbed], components: [actionRow], files: files });

    await interaction.reply({
        embeds: [successEmbed(interaction, 'Fréquence modifiée')],
        flags: [MessageFlags.Ephemeral],
    });
}
