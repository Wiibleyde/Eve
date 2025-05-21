import { ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder, type GuildTextBasedChannel } from 'discord.js';
import { prisma } from '../../utils/database';
import type { StreamData, TwitchUser } from '../../utils/stream/twitch';
import { client } from '../bot';
import { logger } from '../..';

// Renamed functions with more descriptive names
export async function handleStreamStarted(streamer: StreamData, userData: TwitchUser): Promise<void> {
    const embed = generateOnlineStreamEmbed(streamer, userData);
    const streamDatas = await prisma.stream.findMany({
        where: { twitchUserId: streamer.user_id },
    });

    if (!streamDatas || streamDatas.length === 0) {
        logger.warn(`No stream data found for user ID: ${streamer.user_id} (Not supposed to happen)`);
        return;
    }

    for (const streamData of streamDatas) {
        const channel = await client.channels.fetch(streamData.channelId);
        if (!channel?.isTextBased()) continue;

        try {
            // Delete old message if it exists
            if (streamData.messageId) {
                try {
                    const oldMessage = await (channel as GuildTextBasedChannel).messages.fetch(streamData.messageId);
                    if (oldMessage) {
                        await oldMessage.delete();
                    }
                } catch (err) {
                    logger.warn(`Message not found for user ID: ${streamer.user_id}`);
                }
            }

            // Send new message with appropriate mentions
            const newMessageId = await sendStreamMessage(channel as GuildTextBasedChannel, embed, streamData.roleId);

            // Update database with new message ID
            await prisma.stream.update({
                where: { uuid: streamData.uuid },
                data: { messageId: newMessageId },
            });
        } catch (error) {
            logger.error(`Failed to update stream message for ${streamer.user_id}:`, error);
        }
    }
}

// Helper function to send stream messages with optional role mentions
async function sendStreamMessage(
    channel: GuildTextBasedChannel,
    embed: EmbedBuilder,
    roleId?: string | null
): Promise<string | null> {
    try {
        if (roleId && channel.guild) {
            const role = await channel.guild.roles.fetch(roleId);
            if (role) {
                const message = await channel.send({ content: `<@&${role.id}>`, embeds: [embed] });
                return message.id;
            } else {
                logger.warn(`Role ${roleId} not found when sending stream message`);
            }
        }

        const button = new ButtonBuilder()
            .setURL(`https://www.twitch.tv/${embed.data.author?.name}`)
            .setLabel('Watch Stream')
            .setStyle(ButtonStyle.Link);

        const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(button);

        const message = await channel.send({ embeds: [embed], components: [actionRow] });
        return message.id;
    } catch (error) {
        logger.error('Error sending stream message:', error);
        return null;
    }
}

export async function handleStreamEnded(userData: TwitchUser): Promise<void> {
    const embed = generateOfflineStreamEmbed(userData);
    const streamDatas = await prisma.stream.findMany({
        where: { twitchUserId: userData.id },
    });
    if (!streamDatas || streamDatas.length === 0) {
        logger.warn(`No stream data found for user ID: ${userData.id} (Not supposed to happen)`);
        return;
    }
    for (const streamData of streamDatas) {
        const channel = await client.channels.fetch(streamData.channelId);
        if (!channel?.isTextBased()) continue;

        try {
            // Update existing message if it exists
            if (streamData.messageId) {
                try {
                    const existingMessage = await (channel as GuildTextBasedChannel).messages.fetch(streamData.messageId);
                    if (existingMessage) {
                        await existingMessage.edit({ embeds: [embed] });
                        continue; // Skip to next stream data if message was updated successfully
                    }
                } catch (err) {
                    logger.warn(`Message not found for user ID: ${userData.id}, creating new message`);
                }
            }

            // Only create a new message if the existing one wasn't found
            const newMessageId = await sendStreamMessage(channel as GuildTextBasedChannel, embed);

            // Update database with new message ID only if we had to create a new message
            await prisma.stream.update({
                where: { uuid: streamData.uuid },
                data: { messageId: newMessageId },
            });
        } catch (error) {
            logger.error(`Failed to update stream message for ${userData.id}:`, error);
        }
    }
}

export async function handleStreamUpdated(streamer: StreamData, userData: TwitchUser): Promise<void> {
    const embed = generateOnlineStreamEmbed(streamer, userData);
    const streamDatas = await prisma.stream.findMany({
        where: { twitchUserId: streamer.user_id },
    });
    
    if (!streamDatas || streamDatas.length === 0) {
        logger.warn(`No stream data found for user ID: ${streamer.user_id} (Not supposed to happen)`);
        return;
    }
    
    for (const streamData of streamDatas) {
        const channel = await client.channels.fetch(streamData.channelId);
        if (!channel?.isTextBased()) continue;

        try {
            // Only update the existing message if it exists
            if (streamData.messageId) {
                try {
                    const existingMessage = await (channel as GuildTextBasedChannel).messages.fetch(streamData.messageId);
                    if (existingMessage) {
                        await existingMessage.edit({ embeds: [embed] });
                    }
                } catch (err) {
                    logger.warn(`Message not found for user ID: ${streamer.user_id} during stream update`);
                }
            }
        } catch (error) {
            logger.error(`Failed to update stream message for ${streamer.user_id}:`, error);
        }
    }
}

export async function handleInitStreams(onlineStreamers: StreamData[], userData: TwitchUser[]): Promise<void> {
    const streamData = await prisma.stream.findMany();

    for (const streamer of streamData) {
        const stream = onlineStreamers.find((s) => s.user_id === streamer.twitchUserId);
        const user = userData.find((u) => u.id === streamer.twitchUserId);

        // Skip if user not found and delete the record
        if (!user) {
            await deleteStreamMessage(streamer);
            await prisma.stream.delete({
                where: { uuid: streamer.uuid },
            });
            continue;
        }

        // Generate appropriate embed based on stream status
        const embed = stream ? generateOnlineStreamEmbed(stream, user) : generateOfflineStreamEmbed(user);

        // Get channel and update or create message
        const channel = await client.channels.fetch(streamer.channelId);
        if (!channel?.isTextBased()) continue;

        try {
            if (streamer.messageId) {
                const message = await (channel as GuildTextBasedChannel).messages
                    .fetch(streamer.messageId)
                    .catch(() => null);

                if (message) {
                    // Check if the embed needs updating by comparing title and status
                    const currentEmbed = message.embeds[0];
                    const needsUpdate =
                        !currentEmbed ||
                        currentEmbed.title !== embed.data.title ||
                        (stream
                            ? currentEmbed.fields?.find((f) => f.name === 'Viewers')?.value !==
                            String(stream.viewer_count)
                            : false);

                    if (needsUpdate) {
                        await message.edit({ embeds: [embed] });
                    }
                    continue;
                }
            }

            // Create new message if we couldn't update existing one
            const button = new ButtonBuilder()
                .setURL(`https://www.twitch.tv/${user.login}`)
                .setLabel('Watch Channel')
                .setStyle(ButtonStyle.Link);

            const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(button);
            
            const message = await (channel as GuildTextBasedChannel).send({
                embeds: [embed],
                components: stream ? [actionRow] : []
            });
            
            await prisma.stream.update({
                where: { uuid: streamer.uuid },
                data: { messageId: message.id },
            });
            
            logger.info(`Created new message for streamer ${user.display_name} (${stream ? 'online' : 'offline'})`);
        } catch (error) {
            logger.error(`Failed to update stream message for ${streamer.twitchUserId}:`, error);
        }
    }
}

// Helper function to delete stream messages
async function deleteStreamMessage(streamer: { channelId: string; messageId?: string | null }): Promise<void> {
    if (!streamer.messageId) return;

    const channel = await client.channels.fetch(streamer.channelId).catch(() => null);
    if (channel?.isTextBased()) {
        const message = await (channel as GuildTextBasedChannel).messages.fetch(streamer.messageId).catch(() => null);
        if (message) await message.delete().catch(() => null);
    }
}

function generateOnlineStreamEmbed(streamer: StreamData, userData: TwitchUser): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle(streamer.title)
        .setDescription(`**${userData.display_name}** est en direct !`)
        .setImage(streamer.thumbnail_url)
        .setThumbnail(`https://static-cdn.jtvnw.net/ttv-boxart/${streamer.game_id}.jpg`)
        .setURL(`https://www.twitch.tv/${userData.login}`)
        .setColor('#9146FF')
        .setTimestamp(new Date(streamer.started_at))
        .addFields([
            {
                name: 'Game',
                value: streamer.game_name,
                inline: true,
            },
            {
                name: 'Viewers',
                value: String(streamer.viewer_count),
                inline: true,
            },
            {
                name: 'Depuis',
                value: `<t:${Math.floor(new Date(streamer.started_at).getTime() / 1000)}:R>`,
                inline: true,
            },
        ])
        .setAuthor({
            name: userData.display_name,
            iconURL: userData.profile_image_url,
            url: `https://www.twitch.tv/${userData.login}`,
        });
}

function generateOfflineStreamEmbed(userData: TwitchUser): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle('Stream hors ligne')
        .setDescription(`**${userData.display_name}** est hors ligne !`)
        .setThumbnail(userData.profile_image_url)
        .setImage(userData.offline_image_url)
        .setURL(`https://www.twitch.tv/${userData.login}`)
        .setColor('#9146FF')
        .setTimestamp()
        .setAuthor({
            name: userData.display_name,
            iconURL: userData.profile_image_url,
            url: `https://www.twitch.tv/${userData.login}`,
        });
}
