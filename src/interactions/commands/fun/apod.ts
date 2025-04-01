import { config } from '@/config';
import { CommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from 'discord.js';

const url = 'https://api.nasa.gov/planetary/apod';

interface ApodResponse {
    copyright: string;
    date: string;
    explanation: string;
    hdurl: string;
    media_type: string;
    service_version: string;
    title: string;
    url: string;
}

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('apod')
    .setDescription('Astronomy Picture Of the Day');

export async function execute(interaction: CommandInteraction): Promise<void> {
    const response = await fetch(`${url}?api_key=${config.NASA_API_KEY}`);
    const data: ApodResponse = await response.json();

    const embed = new EmbedBuilder()
        .setTitle(data.title)
        .setImage(data.hdurl)
        .setFooter({ text: `Eve – ©${data.copyright}`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();

    const secondEmbed = new EmbedBuilder()
        .setTitle('Explanation')
        .setDescription(data.explanation)
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();

    await interaction.reply({ embeds: [embed, secondEmbed], flags: [MessageFlags.Ephemeral] });
}
