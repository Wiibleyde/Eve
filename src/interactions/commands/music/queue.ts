import { config } from '@/config';
import { errorEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { useQueue } from 'discord-player';
import { CommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('queue')
    .setDescription("[Musique] Afficher la file d'attente");

export async function execute(interaction: CommandInteraction) {
    if (config.MUSIC_MODULE !== true) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Module de musique désactivé.'))],
            ephemeral: true,
        });
        return;
    }
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });
    if (!queue.tracks.toArray()[0])
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const methods = ['', '🔁', '🔂'];
    const songs = queue.tracks.size;
    const nextSongs = songs > 5 ? `et ${songs - 5} autres chansons` : `Dans la file d'attente: ${songs} chansons`;
    const tracks = queue.tracks
        .map((track, i) => `**${i + 1}.** [${track.title}](${track.url}) | ${track.author}`)
        .slice(0, 5)
        .join(backSpace);
    const embed = new EmbedBuilder()
        .setTitle("File d'attente")
        .setDescription(`${methods[queue.repeatMode]} ${nextSongs}${backSpace}${backSpace}${tracks}`)
        .setColor('Blue')
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user?.displayAvatarURL() })
        .setTimestamp();

    await interaction.reply({ embeds: [embed], flags: [MessageFlags.Ephemeral] });
}
