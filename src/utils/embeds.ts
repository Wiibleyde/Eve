import {
    EmbedBuilder,
    CacheType,
    Interaction,
} from 'discord.js';

/**
 * Creates an error embed message for a Discord interaction.
 *
 * @param interaction - The interaction that triggered the error. Can be a CommandInteraction, ButtonInteraction, or ModalSubmitInteraction.
 * @param error - The error object containing the error message to be displayed.
 * @returns An EmbedBuilder instance representing the error message.
 */
export function errorEmbed(
    interaction: Interaction<CacheType>,
    error: Error
): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle('Erreur :warning:')
        .setDescription(error.message)
        .setColor(0xff0000)
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();
}

/**
 * Creates an embed message indicating a successful operation.
 *
 * @param interaction - The interaction object that triggered the command, button, or modal submit.
 * @param message - The success message to be displayed in the embed.
 * @returns An instance of `EmbedBuilder` configured with a success message.
 */
export function successEmbed(
    interaction: Interaction<CacheType>,
    message: string
): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle('Succès !')
        .setDescription(message)
        .setColor(0x00ff00)
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();
}
