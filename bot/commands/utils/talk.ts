import { CommandInteraction, MessageFlags, SlashCommandBuilder, TextChannel, User } from "discord.js";
import type { ICommand } from "../command";
import { errorEmbedGenerator, successEmbedGenerator } from "../../utils/embeds";
import { logger } from "../../..";

export const talk: ICommand = {
    data: new SlashCommandBuilder()
        .setName("talk")
        .setDescription("Make the bot talk")
        .addStringOption(option =>
            option.setName("message")
                .setDescription("The message to send")
                .setRequired(true)
        )
        .addUserOption(option =>
            option.setName("mp")
                .setDescription("The user to send the message to")
        ),
    execute: async (interaction: CommandInteraction) => {
        // Defer reply immediately to acknowledge the interaction
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

        const message = interaction.options.get('message')?.value as string;
        const member = interaction.options.get('mp')?.user as User | null;

        if (member) {
            try {
                await member.send(message);
            } catch (error) {
                logger.error(error);
                await interaction.editReply({
                    embeds: [errorEmbedGenerator("Impossible d'envoyer le message")],
                });
                return;
            }
        } else {
            const channel = interaction.channel as TextChannel;
            if (!channel) {
                await interaction.editReply({
                    embeds: [errorEmbedGenerator('Impossible de trouver le salon de discussion')],
                });
                return;
            }
            await channel.send(message);
        }
        await interaction.editReply({ embeds: [successEmbedGenerator('Message envoyé')] });

    }
};