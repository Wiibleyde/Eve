import { config } from '@/config';
import { logger } from '@/index';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { useMainPlayer, useQueue } from 'discord-player';
import { CommandInteraction, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('syncedlyrics')
    .setDescription('[Musique] Afficher les paroles synchronisées de la musique en cours de lecture');

export async function execute(interaction: CommandInteraction) {
    if (config.MUSIC_MODULE !== true) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Module de musique désactivé."))],
            ephemeral: true,
        });
        return;
    }
    const player = useMainPlayer();
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const metadataThread = (queue.metadata as { lyricsThread: any }).lyricsThread;

    if (metadataThread && !metadataThread.archived)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Les paroles synchronisées sont déjà affichées.'))],
            ephemeral: true,
        });

    const result = await player.lyrics
        .search({
            q: queue.currentTrack?.title,
        })
        .catch(async (error) => {
            logger.error(error);
            return await interaction.reply({ embeds: [errorEmbed(interaction, error)], ephemeral: true });
        });

    const lyrics = Array.isArray(result) ? result[0] : null;

    if (!lyrics) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune parole trouvée pour cette musique.'))],
            ephemeral: true,
        });
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const thread = await (queue.metadata as { channel: any }).channel.threads.create({
        name: `Paroles synchronisées de ${queue.currentTrack?.title.slice(0, 50)}`,
    });

    queue.setMetadata({
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        channel: (queue.metadata as { channel: any }).channel,
        lyricsThread: thread,
    });

    if (!lyrics) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune parole trouvée pour cette musique.'))],
            ephemeral: true,
        });
    }

    const syncedlyrics = queue?.syncedLyrics(lyrics);
    syncedlyrics.onChange(async (lyrics) => {
        await thread.send({ content: lyrics });
    });

    syncedlyrics?.subscribe();

    await interaction.reply({
        embeds: [successEmbed(interaction, 'Paroles synchronisées affichées dans <#' + thread.id + '>')],
        ephemeral: true,
    });
}
