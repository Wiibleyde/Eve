import { CommandInteraction, EmbedBuilder, MessageFlags, PermissionFlagsBits, SlashCommandBuilder } from 'discord.js';
import { prisma } from '@/utils/database';
import { errorEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('logs')
    .setDescription('Afficher les derniers logs du bot');

/**
 * Executes the command to fetch and display the latest logs.
 *
 * @param interaction - The command interaction instance.
 * @returns A promise that resolves when the command execution is complete.
 *
 * @remarks
 * This function first defers the reply to the interaction and checks if the user has the required permissions.
 * If the user lacks the necessary permissions, an error message is sent.
 * Otherwise, it fetches the latest logs from the database, constructs an embed message with the log details,
 * and edits the reply to include the embed.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [PermissionFlagsBits.Administrator], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission d'utiliser cette commande."))],
        });
        return;
    }
    const logs = await prisma.logs.findMany({
        take: 5,
        orderBy: {
            createdAt: 'desc',
        },
        include: {
            level: true,
        },
    });

    const embed = new EmbedBuilder()
        .setTitle('Logs')
        .setDescription('Les derniers logs du bot')
        .setColor(0xffffff)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    logs.forEach((log) => {
        embed.addFields({ name: `Niveau: ${log.level.name}`, value: log.message, inline: false });
    });

    await interaction.editReply({ embeds: [embed] });
}
