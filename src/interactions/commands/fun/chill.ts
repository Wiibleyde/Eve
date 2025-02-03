import { CommandInteraction, EmbedBuilder, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('chill')
    .setDescription('Affiche un gif du mec chill');

export async function execute(interaction: CommandInteraction): Promise<void> {
    const embed = new EmbedBuilder()
        .setTitle('Chill')
        .setDescription('Gif du mec chill')
        .setImage('https://media1.tenor.com/m/Jo0PbgBIZzAAAAAd/chill-guy-my-new-character.gif')
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();

    await interaction.reply({ embeds: [embed], ephemeral: false });
}
