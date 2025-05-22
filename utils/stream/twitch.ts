import axios from 'axios';
import { config } from '../config';
import { logger } from '../..';
import { prisma } from '../database';
import { handleInitStreams, handleStreamEnded, handleStreamStarted, handleStreamUpdated } from '../../bot/utils/stream';
import { client } from '../../bot/bot';
import type { GuildTextBasedChannel } from 'discord.js';

interface OAuthResponse {
    access_token: string;
    expires_in: number;
    token_type: string;
}

export interface TwitchUser {
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

interface TwitchUserResponse {
    data: TwitchUser[];
}

export interface StreamData {
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

interface StreamPagination {
    cursor: string;
}

interface StreamResponse {
    data: StreamData[];
    pagination: StreamPagination;
}

let oauthToken: string | null = null;
let expirationTime: number | null = null;

const oldOnlineStreamers: StreamData[] = [];
const oldUserData: TwitchUser[] = [];

/**
 * Refreshes the OAuth token asynchronously if needed
 * @returns A promise that resolves to the current OAuth token
 */
async function refreshOauthToken(): Promise<string> {
    if (oauthToken && expirationTime && Date.now() < expirationTime) {
        return oauthToken;
    }

    const url = 'https://id.twitch.tv/oauth2/token';

    try {
        const response = await axios.post(url, null, {
            params: {
                client_id: config.TWITCH_CLIENT_ID,
                client_secret: config.TWITCH_CLIENT_SECRET,
                grant_type: 'client_credentials',
            },
        });

        const data: OAuthResponse = response.data;
        oauthToken = data.access_token;
        expirationTime = Date.now() + data.expires_in * 1000;
        logger.debug('Twitch OAuth token refreshed successfully');

        return oauthToken;
    } catch (error) {
        logger.error('Error refreshing Twitch OAuth token:', error);
        throw new Error('Failed to refresh Twitch OAuth token');
    }
}

export async function getUserIdByLogin(login: string): Promise<string | null> {
    const token = await refreshOauthToken();
    const url = `https://api.twitch.tv/helix/users?login=${login}`;
    const headers = {
        Authorization: `Bearer ${token}`,
        'Client-Id': config.TWITCH_CLIENT_ID,
    };
    try {
        const response = await axios.get<TwitchUserResponse>(url, { headers });
        const data = response.data.data;
        if (data.length === 0 || !data[0]) {
            logger.warn(`No Twitch user found for login: ${login}`);
            return null;
        }
        return data[0].id;
    } catch (error) {
        logger.error(`Error fetching Twitch user ID for login ${login}:`, error);
        throw new Error(`Failed to fetch Twitch user ID for login ${login}`);
    }
}

async function getAccountsData(): Promise<TwitchUser[]> {
    const token = await refreshOauthToken();
    const params = await generateDataUrlParams();

    // If we have no streamers to check, return empty array
    if (!params) {
        return [];
    }

    const url = `https://api.twitch.tv/helix/users?${params}`;
    const headers = {
        Authorization: `Bearer ${token}`,
        'Client-Id': config.TWITCH_CLIENT_ID,
    };
    try {
        const response = await axios.get<TwitchUserResponse>(url, { headers });
        const data = response.data.data;
        if (data.length === 0) {
            return [];
        }
        return data.map((user) => ({
            id: user.id,
            login: user.login,
            display_name: user.display_name,
            type: user.type,
            broadcaster_type: user.broadcaster_type,
            description: user.description,
            profile_image_url: user.profile_image_url,
            offline_image_url: user.offline_image_url,
            view_count: user.view_count,
            created_at: user.created_at,
            email: user.email,
        }));
    } catch (error) {
        logger.error('Error fetching Twitch accounts data:', error);
        throw new Error('Failed to fetch Twitch accounts data');
    }
}

async function getStreamsData(): Promise<StreamData[]> {
    const token = await refreshOauthToken();
    const params = await generateOnlineUrlParams();

    // If we have no streamers to check, return empty array
    if (!params) {
        return [];
    }

    const url = `https://api.twitch.tv/helix/streams?${params}`;
    const headers = {
        Authorization: `Bearer ${token}`,
        'Client-Id': config.TWITCH_CLIENT_ID,
    };
    try {
        const response = await axios.get<StreamResponse>(url, { headers });
        const data = response.data.data;
        if (data.length === 0) {
            return [];
        }
        return data;
    } catch (error) {
        logger.error('Error fetching Twitch streams data:', error);
        throw new Error('Failed to fetch Twitch streams data');
    }
}

async function getStreamers(): Promise<string[]> {
    const streamers = await prisma.stream.findMany({
        select: {
            twitchUserId: true,
        },
    });

    return Array.from(new Set(streamers.map((streamer) => streamer.twitchUserId)));
}

async function generateOnlineUrlParams(): Promise<string> {
    const uniqueStreamers = await getStreamers();
    if (uniqueStreamers.length === 0) {
        logger.warn('No streamers found in database for online check');
        return '';
    }
    const limitedStreamers = uniqueStreamers.slice(0, 100);
    return limitedStreamers.map((streamer) => `user_id=${streamer}`).join('&');
}

async function generateDataUrlParams(): Promise<string> {
    const uniqueStreamers = await getStreamers();
    if (uniqueStreamers.length === 0) {
        logger.warn('No streamers found in database for user data check');
        return '';
    }
    const limitedStreamers = uniqueStreamers.slice(0, 100);
    return limitedStreamers.map((streamer) => `id=${streamer}`).join('&');
}

export async function checkStreamsUpdate(): Promise<void> {
    try {
        // Check if we have streamers to monitor before making API calls
        const streamers = await getStreamers();
        if (streamers.length === 0) {
            logger.info('No streamers to monitor in database, skipping update check');
            return;
        }

        // Get streams data with validation
        let onlineStreamers: StreamData[] = [];
        try {
            onlineStreamers = await getStreamsData();
        } catch (error) {
            logger.error('Failed to get streams data, skipping this update cycle:', error);
        }

        // Get user data with validation
        let userData: TwitchUser[] = [];
        try {
            userData = await getAccountsData();
        } catch (error) {
            logger.error('Failed to get account data, continuing with limited functionality:', error);
        }

        // Find streamers who just went online
        const newOnlineStreamers = onlineStreamers.filter((streamer) => {
            return !oldOnlineStreamers.some((oldStreamer) => oldStreamer.id === streamer.id);
        });

        // Find streamers who changed their stream details
        const changedStreamers = onlineStreamers.filter((streamer) => {
            const oldStreamer = oldOnlineStreamers.find((old) => old.id === streamer.id);
            if (!oldStreamer) return false;

            return (
                oldStreamer.game_id !== streamer.game_id ||
                oldStreamer.type !== streamer.type ||
                oldStreamer.title !== streamer.title ||
                oldStreamer.viewer_count !== streamer.viewer_count
            );
        });

        // Find streamers who went offline
        const offlineStreamers = oldOnlineStreamers.filter((oldStreamer) => {
            return !onlineStreamers.some((streamer) => streamer.id === oldStreamer.id);
        });

        // Process new online streamers
        for (const streamer of newOnlineStreamers) {
            const userInfo = userData.find((user) => user.id === streamer.user_id);
            if (userInfo) {
                handleStreamStarted(streamer, userInfo);
            }
        }

        // Process changed streams
        for (const streamer of changedStreamers) {
            const userInfo = userData.find((user) => user.id === streamer.user_id);
            if (userInfo) {
                handleStreamUpdated(streamer, userInfo);
            }
        }

        // Process offline streamers
        for (const streamer of offlineStreamers) {
            const userInfo = oldUserData.find((user) => user.id === streamer.user_id);
            if (userInfo) {
                handleStreamEnded(userInfo);
            } else {
                logger.warn(`No user data found for offline streamer: ${streamer.user_name} (${streamer.user_id})`);
            }
        }

        // Update cached data for next check
        oldOnlineStreamers.length = 0;
        oldOnlineStreamers.push(...onlineStreamers);

        // Update user data cache
        oldUserData.length = 0;
        oldUserData.push(...userData);

        logger.debug(
            `Twitch stream check completed: ${newOnlineStreamers.length} new, ${changedStreamers.length} changed, ${offlineStreamers.length} offline`
        );
    } catch (error) {
        logger.error('Error checking Twitch streams update:', error);
    }
}

export async function initStreamsUpdate(): Promise<void> {
    try {
        oldOnlineStreamers.push(...(await getStreamsData()));
        oldUserData.push(...(await getAccountsData()));
        handleInitStreams(oldOnlineStreamers, oldUserData);
    } catch (error) {
        logger.error('Error initializing Twitch stream update:', error);
    }
}

/**
 * Initializes a single streamer update
 * @param streamerId The Twitch user ID of the streamer to initialize
 * @param forceNotification Whether to force sending a notification even if the streamer is already in the cache
 */
export async function initSingleStreamUpdate(streamerId: string, forceNotification: boolean = false): Promise<void> {
    try {
        const token = await refreshOauthToken();

        // Get user data
        const userUrl = `https://api.twitch.tv/helix/users?id=${streamerId}`;
        const headers = {
            Authorization: `Bearer ${token}`,
            'Client-Id': config.TWITCH_CLIENT_ID,
        };

        const userResponse = await axios.get<TwitchUserResponse>(userUrl, { headers });
        const userData = userResponse.data.data;

        if (userData.length === 0 || !userData[0]) {
            logger.warn(`No Twitch user found for ID: ${streamerId}`);
            return;
        }

        // Get stream data
        const streamUrl = `https://api.twitch.tv/helix/streams?user_id=${streamerId}`;
        const streamResponse = await axios.get<StreamResponse>(streamUrl, { headers });
        const streamData = streamResponse.data.data;

        // Update caches
        const userInfo = userData[0];
        logger.debug(
            `Processing initSingleStreamUpdate for ${userInfo.display_name} (${streamerId}), force: ${forceNotification}`
        );

        // Add or update user in cache
        const userIndex = oldUserData.findIndex((user) => user.id === userInfo.id);
        if (userIndex >= 0) {
            oldUserData[userIndex] = userInfo;
        } else {
            oldUserData.push(userInfo);
        }

        // Process stream data
        if (streamData.length > 0) {
            const streamInfo = streamData[0];
            // Check if stream is already in cache
            const existingStreamIndex = oldOnlineStreamers.findIndex((stream) => stream.user_id === streamerId);

            if (existingStreamIndex >= 0) {
                // Update existing stream data
                if (streamInfo) {
                    oldOnlineStreamers[existingStreamIndex] = streamInfo;
                }
            } else {
                // Add new stream data to cache
                if (streamInfo) {
                    oldOnlineStreamers.push(streamInfo);
                }
            }

            // If the streamer is online, always use handleStreamStarted
            // This will create a proper stream notification with all the formatting
            if (streamInfo && (forceNotification || existingStreamIndex < 0)) {
                logger.debug(`Sending stream notification for ${userInfo.display_name}`);
                await handleStreamStarted(streamInfo, userInfo);
            }
        } else {
            // Streamer is offline, create an offline embed message instead
            logger.debug(`Streamer ${userInfo.display_name} is offline, sending offline notification`);

            // Use the handleStreamEnded function to create an offline message
            if (forceNotification) {
                await handleStreamEnded(userInfo);
            }
        }

        logger.debug(`Single streamer initialization completed for: ${userInfo.display_name} (${streamerId})`);
    } catch (error) {
        logger.error(`Error initializing single streamer (${streamerId}):`, error);
    }
}

/**
 * Removes a streamer from the caches
 * @param streamerId The Twitch user ID of the streamer to remove
 */
export async function removeStreamFromCache(streamerId: string): Promise<void> {
    try {
        // Retrieve streams to delete their messages before they're removed from the database
        const streamsToRemove = await prisma.stream.findMany({
            where: {
                twitchUserId: streamerId,
            },
        });

        // Delete associated messages for each stream
        for (const stream of streamsToRemove) {
            try {
                // Log message details for debugging
                logger.debug(
                    `Attempting to delete message for stream ${stream.uuid} in channel ${stream.channelId}, messageId: ${stream.messageId || 'none'}`
                );

                if (stream.messageId) {
                    // Force delete the message directly instead of updating to offline
                    const channel = await client.channels.fetch(stream.channelId).catch((err) => {
                        logger.warn(`Could not fetch channel ${stream.channelId}: ${err.message}`);
                        return null;
                    });

                    if (channel?.isTextBased()) {
                        try {
                            const message = await (channel as GuildTextBasedChannel).messages.fetch(stream.messageId);
                            if (message) {
                                await message.delete();
                                logger.debug(
                                    `Successfully deleted message ${stream.messageId} from channel ${stream.channelId}`
                                );
                            }
                        } catch (err) {
                            logger.warn(
                                `Could not fetch or delete message ${stream.messageId}: ${err instanceof Error ? err.message : String(err)}`
                            );
                        }
                    }
                }

                logger.debug(`Finished processing message for stream ${stream.uuid}`);
            } catch (error) {
                logger.error(`Failed to delete message for stream ${stream.uuid}:`, error);
            }
        }

        // Check if this is the last reference to this streamer
        const remainingStreamerRefs = await prisma.stream.count({
            where: {
                twitchUserId: streamerId,
            },
        });

        // Only remove from cache if no other guild is using this streamer
        if (remainingStreamerRefs === 0) {
            // Find and remove from oldOnlineStreamers
            const streamIndex = oldOnlineStreamers.findIndex((stream) => stream.user_id === streamerId);
            if (streamIndex >= 0) {
                oldOnlineStreamers.splice(streamIndex, 1);
            }

            // Find and remove from oldUserData
            const userIndex = oldUserData.findIndex((user) => user.id === streamerId);
            if (userIndex >= 0) {
                const userInfo = oldUserData[userIndex];
                oldUserData.splice(userIndex, 1);

                if (userInfo) {
                    logger.debug(`Streamer removed from cache: ${userInfo.display_name} (${streamerId})`);
                } else {
                    logger.debug(`Streamer removed from cache: unknown display_name (${streamerId})`);
                }
            }
        }
    } catch (error) {
        logger.error(`Error removing streamer from cache (${streamerId}):`, error);
    }
}
