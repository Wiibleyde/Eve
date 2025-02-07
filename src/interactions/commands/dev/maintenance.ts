import { CommandInteraction, SlashCommandBuilder, EmbedBuilder, PermissionFlagsBits, MessageFlags } from 'discord.js';
import { errorEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';

export let maintenance: boolean = false;

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('maintenance')
    .setDescription('Passer le bot en mode maintenance');

/**
 * Executes the maintenance command, toggling the maintenance mode of the bot.
 *
 * @param interaction - The command interaction that triggered this execution.
 * @returns A promise that resolves when the command execution is complete.
 *
 * The command performs the following steps:
 * 1. Defers the reply to the interaction, making it ephemeral and fetching the reply.
 * 2. Checks if the user has the necessary permissions (Administrator) to execute the command.
 *    - If not, it edits the reply with an error message and exits.
 * 3. Toggles the maintenance mode of the bot.
 * 4. Creates an embed message indicating the new state of the bot (maintenance or normal).
 * 5. Edits the reply with the embed message.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [PermissionFlagsBits.Administrator], true))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission d'utiliser cette commande."))],
        });
        return;
    }
    maintenance = !maintenance;
    const color: number = maintenance ? 0xff0000 : 0x00ff00;
    const embed = new EmbedBuilder()
        .setTitle('Maintenance')
        .setDescription(`Le bot est maintenant en mode ${maintenance ? 'maintenance' : 'normal'}`)
        .setColor(color)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    await interaction.editReply({ embeds: [embed] });
}
