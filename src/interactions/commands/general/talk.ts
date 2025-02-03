import {
    CommandInteraction,
    PermissionFlagsBits,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
    TextChannel,
} from 'discord.js';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { logger } from '@/index';
import { hasPermission } from '@/utils/permissionTester';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('talk')
    .setDescription('Parler avec en utilisant bot')
    .addStringOption((option) => option.setName('message').setDescription('Le message à envoyer').setRequired(true))
    .addStringOption((option) => option.setName('mp').setDescription('ID de la personne').setRequired(false));

/**
 * Executes the command to send a message either to a specific user or to the current channel.
 *
 * @param interaction - The command interaction object containing the options and context.
 * @returns A promise that resolves when the command execution is complete.
 *
 * The function performs the following steps:
 * 1. Defers the reply to the interaction to allow for asynchronous processing.
 * 2. Retrieves the message and member (if any) from the interaction options.
 * 3. If a member is specified, attempts to fetch the user and send them the message via direct message.
 * 4. If no member is specified, sends the message to the current channel.
 * 5. Handles errors by editing the reply with an appropriate error message.
 * 6. Finally, edits the reply to indicate that the message was sent successfully.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ ephemeral: true, fetchReply: true });

    if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageMessages], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission d'utiliser cette commande."))],
        });
        return;
    }

    const message = interaction.options.get('message')?.value as string;
    const member = interaction.options.get('mp')?.value as string;

    if (member) {
        const user = await interaction.client.users.fetch(member);
        if (!user) {
            await interaction.editReply({
                embeds: [errorEmbed(interaction, new Error("Impossible de trouver l'utilisateur"))],
            });
            return;
        }
        try {
            await user.send(message);
        } catch (error) {
            logger.error(error);
            await interaction.editReply({
                embeds: [errorEmbed(interaction, new Error("Impossible d'envoyer le message"))],
            });
            return;
        }
    } else {
        const channel = interaction.channel as TextChannel;
        if (!channel) {
            await interaction.editReply({
                embeds: [errorEmbed(interaction, new Error('Impossible de trouver le salon de discussion'))],
            });
            return;
        }
        await channel.send(message);
    }
    await interaction.editReply({ embeds: [successEmbed(interaction, 'Message envoyé')] });
}
