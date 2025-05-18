import { CommandInteraction, MessageFlags, SlashCommandBuilder } from "discord.js";
import type { ICommand } from "../command";
import { basicEmbedGenerator } from "../../utils/embeds";

const PING_THRESHOLD = {
    VERY_GOOD: 50,
    GOOD: 100,
    CORRECT: 150,
    WEAK: 200,
    BAD: 500,
};

export const ping: ICommand = {
    data: new SlashCommandBuilder()
        .setName("ping")
        .setDescription("Replies with Pong!"),
    execute: async (interaction: CommandInteraction) => {
        const firstResponse = await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
        let status: string;
        let color: number;

        const memoryData: NodeJS.MemoryUsage = process.memoryUsage();
        const ping: number = interaction.client.ws.ping;

        if (!firstResponse || ping <= 0) {
            status = 'Surprenant ! ⚫';
            color = 0x000000;
        } else if (ping < PING_THRESHOLD.VERY_GOOD) {
            status = 'Très bon 🟢';
            color = 0x00ff00;
        } else if (ping < PING_THRESHOLD.GOOD) {
            status = 'Bon 🟢';
            color = 0x00ff00;
        } else if (ping < PING_THRESHOLD.CORRECT) {
            status = 'Correct 🟡';
            color = 0x00ff00;
        } else if (ping < PING_THRESHOLD.WEAK) {
            status = 'Faible 🟠';
            color = 0xffa500;
        } else if (ping < PING_THRESHOLD.BAD) {
            status = 'Mauvais 🔴';
            color = 0xff0000;
        } else {
            status = 'Très mauvais 🔴';
            color = 0xff0000;
        }

        const pingEmbed = basicEmbedGenerator()
        pingEmbed.setTitle('Ping')
        pingEmbed.setDescription('Le bot est connecté et répond')
        pingEmbed.addFields(
            { name: 'Ping', value: `${ping}ms / ${status}`, inline: true },
            { name: 'Mémoire', value: `${(memoryData.heapUsed / 1024 / 1024).toFixed(2)} MB`, inline: true },
            { name: 'Uptime', value: `${(process.uptime() / 60).toFixed(2)} minutes`, inline: true }
        )
        pingEmbed.setColor(color)

        await interaction.editReply({ embeds: [pingEmbed] });

    }
};