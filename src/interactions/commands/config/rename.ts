import {
    CommandInteraction,
    SlashCommandBuilder,
    PermissionFlagsBits,
    InteractionContextType,
    SlashCommandOptionsOnlyBuilder,
    MessageFlags,
} from 'discord.js';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('rename')
    .setDescription('Renommer le bot')
    .addStringOption((option) => option.setName('nom').setDescription('Le nouveau nom du bot').setRequired(false))
    .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]);

/**
 * Executes the rename command to change the bot's nickname.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves when the command execution is complete.
 *
 * The function performs the following steps:
 * 1. Defers the reply to the interaction.
 * 2. Checks if the user has the required permissions to manage channels.
 * 3. If the user lacks permissions, sends an error message and returns.
 * 4. Retrieves the new name from the interaction options.
 * 5. If the new name is empty, resets the bot's nickname and sends a success message.
 * 6. If the new name exceeds 32 characters, sends an error message.
 * 7. If the new name is the same as the current username, sends an error message.
 * 8. Otherwise, sets the bot's nickname to the new name and sends a success message.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de changer le nom du bot"))],
        });
        return;
    }
    const newName = interaction.options.get('nom')?.value as string;
    if (!newName || newName.length === 0) {
        interaction.guild?.members.cache.get(interaction.client.user.id)?.setNickname('');
        await interaction.editReply({ embeds: [successEmbed(interaction, `Le nom du bot a été réinitialisé`)] });
    } else if (newName.length > 32) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error('Le nom du bot ne doit pas dépasser 32 caractères'))],
        });
    } else if (newName === interaction.client.user.username) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error('Le nom du bot ne peut pas être le même que le nom actuel'))],
        });
    } else {
        interaction.guild?.members.cache.get(interaction.client.user.id)?.setNickname(newName);
        await interaction.editReply({
            embeds: [successEmbed(interaction, `Le nom du bot a été changé en ${newName}`)],
        });
    }
}
