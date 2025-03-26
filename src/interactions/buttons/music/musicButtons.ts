import { logger, player } from '@/index';
import { back } from '@/interactions/commands/music/back';
import { skip } from '@/interactions/commands/music/skip';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { getAssociatedMusic } from '@/utils/intelligence';
import { waitTime } from '@/utils/utils';
import { QueryType, QueueRepeatMode, useQueue } from 'discord-player';
import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonInteraction,
    ButtonStyle,
    EmbedBuilder,
    GuildMember,
    MessageActionRowComponentBuilder,
} from 'discord.js';

export function backButton(interaction: ButtonInteraction) {
    back(interaction);
}

export async function loopButton(interaction: ButtonInteraction) {
    const methods = ['disabled', 'track', 'queue'];
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    if (queue.repeatMode === QueueRepeatMode.QUEUE) {
        queue.setRepeatMode(QueueRepeatMode.OFF);
    } else {
        queue.setRepeatMode((queue.repeatMode + 1) as QueueRepeatMode);
    }

    return await interaction.reply({
        embeds: [successEmbed(interaction, `Boucle ${methods[queue.repeatMode]}`)],
        ephemeral: true,
    });
}

export async function resumeAndPauseButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const resumed = queue.node.resume();

    if (!resumed) {
        queue.node.pause();
        await interaction.reply({ embeds: [successEmbed(interaction, 'Musique mise en pause')] });
        await waitTime(5000);
        await interaction.deleteReply();
        return;
    }

    await interaction.reply({ embeds: [successEmbed(interaction, 'Musique reprise')] });
    await waitTime(5000);
    await interaction.deleteReply();
}

export function skipButton(interaction: ButtonInteraction) {
    skip(interaction);
}

export async function iaButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const nowPlaying = queue.currentTrack;

    if (nowPlaying?.cleanTitle && nowPlaying?.author) {
        const musics = await getAssociatedMusic(nowPlaying.cleanTitle, nowPlaying.author);

        const embed = new EmbedBuilder()
            .setTitle("L'IA vous propose")
            .setDescription("Les musiques suivantes sont proposées par l'IA elles peuvent ne pas être exactes.")
            .setColor('Blue')
            .setTimestamp()
            .setFooter({
                text: `Eve – Toujours prête à vous aider.`,
                iconURL: interaction.client.user.displayAvatarURL(),
            });

        const fields = musics.map((music, index) => {
            return {
                name: `${index + 1}- ${music.title}`,
                value: `Auteur: ${music.author}`,
            };
        });

        embed.addFields(fields);

        const button = function (index: number) {
            return new ButtonBuilder()
                .setLabel('Ajouter n°' + (index + 1))
                .setCustomId('addTrackButton--' + index)
                .setStyle(ButtonStyle.Primary);
        };

        const row = new ActionRowBuilder<MessageActionRowComponentBuilder>().addComponents(
            musics.map((_, index) => button(index))
        );

        await interaction.reply({ embeds: [embed], components: [row], ephemeral: true });
    } else {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de récupérer les informations de la musique.'))],
            ephemeral: true,
        });
    }
}

export async function addTrackButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const nowPlaying = queue.currentTrack;

    if (nowPlaying?.cleanTitle && nowPlaying?.author) {
        const musics = await getAssociatedMusic(nowPlaying.cleanTitle, nowPlaying.author);

        const index = parseInt(interaction.customId.split('--')[1]);

        if (musics[index]) {
            const res = await player.search(`${musics[index].title} ${musics[index].author}`, {
                requestedBy: interaction.user,
                searchEngine: QueryType.AUTO,
            });

            if (!res?.tracks.length) {
                return await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error('Aucun résultat trouvé.'))],
                    ephemeral: true,
                });
            }

            const userVoiceChannel = (interaction.member as GuildMember)?.voice.channel;
            if (!userVoiceChannel) {
                return await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error('Vous devez être dans un salon vocal.'))],
                    ephemeral: true,
                });
            }

            try {
                logger.debug(`Adding track ${musics[index].title} ${musics[index].author} to queue`);
                const { track } = await player.play(
                    userVoiceChannel,
                    `${musics[index].title} ${musics[index].author}`,
                    {
                        nodeOptions: {
                            metadata: {
                                channel: interaction.channel,
                            },
                            volume: 100,
                            leaveOnEmpty: true,
                            leaveOnEmptyCooldown: 60000,
                            leaveOnEnd: true,
                            leaveOnEndCooldown: 60000,
                        },
                    }
                );

                await interaction.reply({
                    embeds: [
                        successEmbed(
                            interaction,
                            `Musique ajoutée à la file d'attente: [${track.title}](${track.url})`
                        ),
                    ],
                });
                await waitTime(5000);
                await interaction.deleteReply();
            } catch (error) {
                logger.error(error);
                await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error('Impossible de jouer la musique.'))],
                    ephemeral: true,
                });
            }
        } else {
            await interaction.reply({
                embeds: [errorEmbed(interaction, new Error('Impossible de récupérer les informations de la musique.'))],
                ephemeral: true,
            });
        }
    } else {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de récupérer les informations de la musique.'))],
            ephemeral: true,
        });
    }
}
