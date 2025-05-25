import { CommandInteraction, MessageFlags, SlashCommandBuilder, TextChannel, User } from 'discord.js';
import type { ICommand } from '../command';
import { errorEmbedGenerator, successEmbedGenerator } from '../../utils/embeds';
import { logger } from '../../..';

export const talk: ICommand = {
    data: new SlashCommandBuilder()
        .setName('talk')
        .setDescription('Make the bot talk')
        .addStringOption((option) => option.setName('message').setDescription('The message to send').setRequired(true))
        .addUserOption((option) => option.setName('mp').setDescription('The user to send the message to')),
    execute: async (interaction: CommandInteraction) => {
        //TODO: Check if the user has the required permissions to execute this command
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

        const message = interaction.options.get('message')?.value as string;
        const targetUser = interaction.options.get('mp')?.user as User | null;

        try {
            if (targetUser) {
                await sendDirectMessage(targetUser, message);
            } else {
                await sendChannelMessage(interaction, message);
            }
            await interaction.editReply({
                embeds: [successEmbedGenerator('Message envoyé')],
            });
        } catch (error) {
            logger.error(error);
            await interaction.editReply({
                embeds: [
                    errorEmbedGenerator(error instanceof Error ? error.message : "Impossible d'envoyer le message"),
                ],
            });
        }
    },
};

/**
 * Send a direct message to a user
 */
async function sendDirectMessage(user: User, message: string): Promise<void> {
    try {
        await user.send(message);
    } catch {
        throw new Error("Impossible d'envoyer le message privé");
    }
}

/**
 * Send a message to the current channel
 */
async function sendChannelMessage(interaction: CommandInteraction, message: string): Promise<void> {
    const channel = interaction.channel as TextChannel;
    if (!channel) {
        throw new Error('Impossible de trouver le salon de discussion');
    }
    await channel.send(message);
}
