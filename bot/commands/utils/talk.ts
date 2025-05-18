import { CommandInteraction, SlashCommandBuilder } from "discord.js";
import type { ICommand } from "../command";

export const talk: ICommand = {
    data: new SlashCommandBuilder()
        .setName("talk")
        .setDescription("Make the bot talk")
        .addStringOption(option =>
            option.setName("message")
                .setDescription("The message to send")
                .setRequired(true)
        ),
    execute: async (interaction: CommandInteraction) => {
        const message = interaction.options.get("message")?.value as string;
        if (message) {
            await interaction.reply(message);
        } else {
            await interaction.reply("Please provide a message.");
        }
    }
};