import { CommandInteraction, SlashCommandBuilder } from "discord.js";
import type { ICommand } from "../command";

export const ping: ICommand = {
    data: new SlashCommandBuilder()
        .setName("ping")
        .setDescription("Replies with Pong!"),
    execute: async (interaction: CommandInteraction) => {
        await interaction.reply("Pong!");
    }
};