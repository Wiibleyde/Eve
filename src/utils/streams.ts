import { client, logger } from '..';
import { prisma } from './database';
import { EmbedBuilder, Message, TextChannel } from 'discord.js';

interface StreamData {
    id: string;
    user_id: string;
    user_name: string;
    game_id: string;
    game_name: string;
    type: string;
    title: string;
    tags: string[];
    viewer_count: number;
    started_at: string;
    language: string;
    thumbnail_url: string;
    tag_ids: string[];
    is_mature: boolean;
}

interface OfflineStreamData {
    id: string;
    login: string;
    display_name: string;
    type: string;
    broadcaster_type: string;
    description: string;
    profile_image_url: string;
    offline_image_url: string;
    view_count: number;
    email?: string;
    created_at: string;
}

function generateEmbed(online: boolean, stream: StreamData | null, offlineData: OfflineStreamData): EmbedBuilder {
    let embed: EmbedBuilder;
    if (online && stream) {
        embed = new EmbedBuilder()
            .setTitle(offlineData.display_name)
            .setDescription(stream.title)
            .setAuthor({
                name: stream.user_name,
                url: `https://twitch.tv/${stream.user_name}`,
                iconURL: offlineData.profile_image_url,
            })
            .setURL(`https://twitch.tv/${stream.user_name}`)
            .setImage(stream.thumbnail_url.replace('{width}', '1280').replace('{height}', '720'))
            .setThumbnail(`https://static-cdn.jtvnw.net/ttv-boxart/${stream.game_id}.jpg`)
            .addFields(
                {
                    name: 'En train de :',
                    value: stream.game_name || 'Inconnu',
                    inline: true,
                },
                {
                    name: 'Démarré :',
                    value: `<t:${Math.floor(new Date(stream.started_at).getTime() / 1000)}:R>`,
                    inline: true,
                }
            )
            .setColor('White')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client?.user?.displayAvatarURL() })
            .setTimestamp();
    } else {
        embed = new EmbedBuilder()
            .setTitle(offlineData.display_name)
            .setDescription('Le stream est hors-ligne.')
            .setAuthor({
                name: offlineData.display_name,
                url: `https://twitch.tv/${offlineData.login}`,
                iconURL: offlineData.profile_image_url,
            })
            .setURL(`https://twitch.tv/${offlineData.login}`)
            .setImage(offlineData.offline_image_url)
            .setColor('DarkerGrey')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client?.user?.displayAvatarURL() })
            .setTimestamp();
    }
    return embed;
}

export async function onStreamOnline(stream: StreamData, offlineData: OfflineStreamData) {
    const databaseStreamData = await prisma.stream.findMany({
        where: {
            twitchChannelName: stream.user_name,
        },
    });
    for (const streamData of databaseStreamData) {
        let mention: string = '';
        if (streamData.roleId) {
            mention = `<@&${streamData.roleId}>`;
        } else {
            mention = '';
        }
        if (!streamData.messageId) {
            const embed = generateEmbed(true, stream, offlineData);
            const channel = (await client.channels.fetch(streamData.channelId)) as TextChannel;
            if (channel) {
                const message = await channel.send({ content: mention, embeds: [embed] });
                await prisma.stream.update({
                    where: {
                        uuid: streamData.uuid,
                    },
                    data: {
                        messageId: message.id,
                    },
                });
            }
        } else if (streamData.messageId) {
            const channel = (await client.channels.fetch(streamData.channelId)) as TextChannel;
            if (channel) {
                let message: Message | null = null;
                try {
                    message = await channel.messages.fetch(streamData.messageId as string);
                } catch (error) {
                    logger.error('Error while fetching message:', error);
                    await prisma.stream.update({
                        where: {
                            uuid: streamData.uuid,
                        },
                        data: {
                            messageId: null,
                        },
                    });
                    return onStreamOnline(stream, offlineData);
                }
                if (message) {
                    if (message.embeds.length > 0) {
                        const currentEmbed = message.embeds[0];
                        if (
                            currentEmbed.description !== stream.title ||
                            !currentEmbed.fields?.some(
                                (field) =>
                                    field.name === 'En train de :' &&
                                    field.value === (stream.game_name || 'Inconnu')
                            )
                        ) {
                            const embed = generateEmbed(true, stream, offlineData);
                            await message.edit({ content: mention, embeds: [embed] });
                        }
                        return;
                    }

                    const embed = generateEmbed(true, stream, offlineData);
                    try {
                        await message.delete();
                    } catch (error) {
                        logger.error('Error while deleting message:', error);
                    }
                    const newMessage = await channel.send({ content: mention, embeds: [embed] });
                    await prisma.stream.update({
                        where: {
                            uuid: streamData.uuid,
                        },
                        data: {
                            messageId: newMessage.id,
                        },
                    });
                }
            }
        }
    }
}

export async function onStreamOffline(offlineData: OfflineStreamData) {
    const databaseStreamData = await prisma.stream.findMany({
        where: {
            twitchChannelName: offlineData.login,
        },
    });
    for (const streamData of databaseStreamData) {
        const channel = (await client.channels.fetch(streamData.channelId)) as TextChannel;
        if (channel) {
            if (streamData.messageId) {
                try {
                    const message = await channel.messages.fetch(streamData.messageId as string);
                    if (message) {
                        if (message.embeds.length > 0 && message.embeds[0].description === 'Le stream est hors-ligne.') {
                            continue;
                        }
                        const embed = generateEmbed(false, null, offlineData);
                        await message.edit({
                            content: null,
                            embeds: [embed],
                        });
                    }
                } catch (error) {
                    logger.error('Error while fetching message for offline update:', error);
                    const embed = generateEmbed(false, null, offlineData);
                    const newMessage = await channel.send({
                        embeds: [embed],
                    });
                    await prisma.stream.update({
                        where: {
                            uuid: streamData.uuid,
                        },
                        data: {
                            messageId: newMessage.id,
                        },
                    });
                }
            } else {
                const embed = generateEmbed(false, null, offlineData);
                const message = await channel.send({
                    embeds: [embed],
                });
                await prisma.stream.update({
                    where: {
                        uuid: streamData.uuid,
                    },
                    data: {
                        messageId: message.id,
                    },
                });
            }
        }
    }
}

export async function removeStream(streamer: string, guildId: string) {
    const databaseStreamData = await prisma.stream.findFirst({
        where: {
            AND: [
                {
                    twitchChannelName: streamer,
                },
                {
                    guildId: guildId,
                },
            ],
        },
    });

    if (!databaseStreamData) {
        return;
    }

    const channel = (await client.channels.fetch(databaseStreamData.channelId)) as TextChannel;
    if (channel) {
        if (databaseStreamData.messageId) {
            const message = await channel.messages.fetch(databaseStreamData.messageId as string);
            if (message) {
                await message.delete();
            }
        }
    }

    await prisma.stream.delete({
        where: {
            uuid: databaseStreamData.uuid,
        },
    });
}
