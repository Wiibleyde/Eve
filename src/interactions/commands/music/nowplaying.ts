import { config, playerConfig } from '@/config';
import { errorEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { useQueue } from 'discord-player';
import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    CommandInteraction,
    EmbedBuilder,
    MessageActionRowComponentBuilder,
    SlashCommandBuilder,
} from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('nowplaying')
    .setDescription('[Musique] Afficher la musique en cours de lecture');

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

    const track = queue.currentTrack;
    if (!track)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });
    const progress = queue.node.createProgressBar();

    let emojiState = playerConfig.enableEmoji;

    const emojis = playerConfig.emojis;

    emojiState = emojis ? true : false;

    const embed = new EmbedBuilder()
        .setTitle('Lecture en cours')
        .setThumbnail(track.thumbnail)
        .setDescription(`[${track.title}](${track.url}) | ${track.author} ${backSpace} <${progress}>`)
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() })
        .setTimestamp();

    const back = new ButtonBuilder()
        .setLabel(emojiState ? emojis.back : 'Back')
        .setCustomId('backButton')
        .setStyle(ButtonStyle.Primary);

    const skip = new ButtonBuilder()
        .setLabel(emojiState ? emojis.skip : 'Skip')
        .setCustomId('skipButton')
        .setStyle(ButtonStyle.Primary);

    const resumepause = new ButtonBuilder()
        .setLabel(emojiState ? emojis.ResumePause : 'Resume/Pause')
        .setCustomId('resumeAndPauseButton')
        .setStyle(ButtonStyle.Danger);

    const loop = new ButtonBuilder()
        .setLabel(emojiState ? emojis.loop : 'Loop')
        .setCustomId('loopButton')
        .setStyle(ButtonStyle.Danger);

    // const lyrics = new ButtonBuilder() // Disabled for now
    //     .setLabel('Lyrics')
    //     .setCustomId('lyricsButton')
    //     .setStyle(ButtonStyle.Secondary)

    const row = new ActionRowBuilder<MessageActionRowComponentBuilder>().addComponents(back, skip, resumepause, loop);
    await interaction.reply({ embeds: [embed], components: [row], ephemeral: true });
}
