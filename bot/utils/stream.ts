import { ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder, type GuildTextBasedChannel } from 'discord.js';
import { prisma } from '../../utils/database';
import type { StreamData, TwitchUser } from '../../utils/stream/twitch';
import { client } from '../bot';
import { logger } from '../..';

// Common helper functions to reduce duplication
async function getStreamDataForUser(userId: string): Promise<{ streamDatas: any[]; success: boolean }> {
    try {
        const streamDatas = await prisma.stream.findMany({
            where: { twitchUserId: userId },
        });

        if (!streamDatas || streamDatas.length === 0) {
            logger.warn(`No stream data found for user ID: ${userId} (Not supposed to happen)`);
            return { streamDatas: [], success: false };
        }

        return { streamDatas, success: true };
    } catch (error) {
        logger.error(`Failed to fetch stream data for ${userId}:`, error);
        return { streamDatas: [], success: false };
    }
}

async function getTextChannel(channelId: string): Promise<GuildTextBasedChannel | null> {
    try {
        const channel = await client.channels.fetch(channelId);
        if (!channel?.isTextBased()) return null;
        return channel as GuildTextBasedChannel;
    } catch (error) {
        logger.warn(`Failed to fetch channel ${channelId}:`, error);
        return null;
    }
}

async function updateOrCreateMessage(
    channel: GuildTextBasedChannel,
    embed: EmbedBuilder,
    existingMessageId?: string | null,
    roleId?: string | null,
    addButton = true,
    streamerName?: string
): Promise<string | null> {
    try {
        // Try to update existing message
        if (existingMessageId) {
            try {
                const existingMessage = await channel.messages.fetch(existingMessageId);
                if (existingMessage) {
                    const components = addButton && streamerName ? [createStreamButton(streamerName)] : [];
                    await existingMessage.edit({ embeds: [embed], components });
                    return existingMessageId;
                }
            } catch (err) {
                logger.warn(`Message ${existingMessageId} not found, creating new message`);
            }
        }

        // Create new message if needed
        return await sendStreamMessage(channel, embed, roleId, addButton, streamerName);
    } catch (error) {
        logger.error('Error updating/creating message:', error);
        return null;
    }
}

function createStreamButton(streamerName: string): ActionRowBuilder<ButtonBuilder> {
    const button = new ButtonBuilder()
        .setURL(`https://www.twitch.tv/${streamerName}`)
        .setLabel('Watch Stream')
        .setStyle(ButtonStyle.Link);

    return new ActionRowBuilder<ButtonBuilder>().addComponents(button);
}

// Helper function to send stream messages with optional role mentions
async function sendStreamMessage(
    channel: GuildTextBasedChannel,
    embed: EmbedBuilder,
    roleId?: string | null,
    addButton = true,
    streamerName?: string
): Promise<string | null> {
    try {
        const content = roleId && channel.guild ? `<@&${roleId}>` : undefined;
        const components = addButton && streamerName ? [createStreamButton(streamerName)] : [];

        const message = await channel.send({
            content,
            embeds: [embed],
            components,
        });

        return message.id;
    } catch (error) {
        logger.error('Error sending stream message:', error);
        return null;
    }
}

// Main functions using the refactored helpers
export async function handleStreamStarted(streamer: StreamData, userData: TwitchUser): Promise<void> {
    const embed = generateOnlineStreamEmbed(streamer, userData);
    const { streamDatas, success } = await getStreamDataForUser(streamer.user_id);
    if (!success) return;

    for (const streamData of streamDatas) {
        const channel = await getTextChannel(streamData.channelId);
        if (!channel) continue;

        try {
            // Delete old message if it exists
            if (streamData.messageId) {
                try {
                    const oldMessage = await channel.messages.fetch(streamData.messageId);
                    if (oldMessage) {
                        await oldMessage.delete();
                    }
                } catch (err) {
                    logger.warn(`Message not found for user ID: ${streamer.user_id}`);
                }
            }

            // Send new message with appropriate mentions
            const newMessageId = await sendStreamMessage(channel, embed, streamData.roleId, true, userData.login);

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

export async function handleStreamEnded(userData: TwitchUser): Promise<void> {
    const embed = generateOfflineStreamEmbed(userData);
    const { streamDatas, success } = await getStreamDataForUser(userData.id);
    if (!success) return;

    for (const streamData of streamDatas) {
        const channel = await getTextChannel(streamData.channelId);
        if (!channel) continue;

        try {
            const newMessageId = await updateOrCreateMessage(
                channel,
                embed,
                streamData.messageId,
                null,
                false,
                userData.login
            );

            // Update database if needed
            if (newMessageId !== streamData.messageId) {
                await prisma.stream.update({
                    where: { uuid: streamData.uuid },
                    data: { messageId: newMessageId },
                });
            }
        } catch (error) {
            logger.error(`Failed to update stream message for ${userData.id}:`, error);
        }
    }
}

export async function handleStreamUpdated(streamer: StreamData, userData: TwitchUser): Promise<void> {
    const embed = generateOnlineStreamEmbed(streamer, userData);
    const { streamDatas, success } = await getStreamDataForUser(streamer.user_id);
    if (!success) return;

    for (const streamData of streamDatas) {
        const channel = await getTextChannel(streamData.channelId);
        if (!channel) continue;

        try {
            // Only update the existing message if it exists
            if (streamData.messageId) {
                await updateOrCreateMessage(channel, embed, streamData.messageId, null, true, userData.login);
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
        const channel = await getTextChannel(streamer.channelId);
        if (!channel) continue;

        try {
            const newMessageId = await updateOrCreateMessage(
                channel,
                embed,
                streamer.messageId,
                null,
                !!stream,
                user.login
            );

            if (newMessageId !== streamer.messageId) {
                await prisma.stream.update({
                    where: { uuid: streamer.uuid },
                    data: { messageId: newMessageId },
                });
            }

            logger.info(`Updated message for streamer ${user.display_name} (${stream ? 'online' : 'offline'})`);
        } catch (error) {
            logger.error(`Failed to update stream message for ${streamer.twitchUserId}:`, error);
        }
    }
}

// Helper function to delete stream messages
export async function deleteStreamMessage(streamer: { channelId: string; messageId?: string | null }): Promise<void> {
    if (!streamer.messageId) {
        logger.debug(`No message ID found for stream in channel ${streamer.channelId}`);
        return;
    }

    const channel = await getTextChannel(streamer.channelId);
    if (!channel) return;

    try {
        logger.debug(`Attempting to fetch message ${streamer.messageId} from channel ${streamer.channelId}`);

        const message = await channel.messages.fetch(streamer.messageId).catch((err) => {
            logger.warn(`Could not fetch message ${streamer.messageId}: ${err.message}`);
            return null;
        });

        if (message) {
            logger.debug(`Deleting message ${streamer.messageId}`);
            await message.delete().catch((err) => {
                logger.warn(`Could not delete message ${streamer.messageId}: ${err.message}`);
            });
            logger.debug(`Successfully deleted message ${streamer.messageId}`);
        } else {
            logger.warn(`Message ${streamer.messageId} not found in channel ${streamer.channelId}`);
        }
    } catch (error) {
        logger.error(`Error in deleteStreamMessage for message ${streamer.messageId}:`, error);
    }
}

/**
 * Generates an embed for when a streamer is online
 * @param streamer - The stream data from Twitch
 * @param userData - The Twitch user data for the streamer
 * @returns An EmbedBuilder with online stream information
 */
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
                value: `<t:${Math.floor(new Date(streamer.started_at).getTime() / 1000)}:R>`, // Uses Discord timestamp format for relative time
                inline: true,
            },
        ])
        .setAuthor({
            name: userData.display_name,
            iconURL: userData.profile_image_url, // Uses streamer's profile image as the author icon
            url: `https://www.twitch.tv/${userData.login}`, // Links to streamer's Twitch channel
        });
}

/**
 * Generates an embed for when a streamer is offline
 * @param userData - The Twitch user data for the streamer
 * @returns An EmbedBuilder with offline stream information
 */
function generateOfflineStreamEmbed(userData: TwitchUser): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle('Stream hors ligne') // "Stream offline" in French
        .setDescription(`**${userData.display_name}** est hors ligne !`) // "{name} is offline!" in French
        .setThumbnail(userData.profile_image_url) // Small profile image in the corner
        .setImage(userData.offline_image_url) // Uses the streamer's custom offline banner if available
        .setURL(`https://www.twitch.tv/${userData.login}`) // Links to streamer's Twitch channel
        .setColor('#9146FF') // Twitch purple color
        .setTimestamp() // Current timestamp when the embed was created
        .setAuthor({
            name: userData.display_name,
            iconURL: userData.profile_image_url,
            url: `https://www.twitch.tv/${userData.login}`,
        });
}
