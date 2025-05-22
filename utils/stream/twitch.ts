import axios from 'axios';
import { config } from '../config';
import { logger } from '../..';
import { prisma } from '../database';
import { handleInitStreams, handleStreamEnded, handleStreamStarted, handleStreamUpdated } from '../../bot/utils/stream';
import { client } from '../../bot/bot';
import type { GuildTextBasedChannel } from 'discord.js';

/**
 * Response structure for OAuth token requests
 */
interface OAuthResponse {
    access_token: string;
    expires_in: number;
    token_type: string;
}

/**
 * Represents a Twitch user's profile information
 */
export interface TwitchUser {
    id: string; // Unique Twitch user ID
    login: string; // Username for the URL (e.g., twitch.tv/login)
    display_name: string; // User's display name which may include capitalization
    type: string; // Account type
    broadcaster_type: string; // Broadcaster classification (partner, affiliate, etc.)
    description: string; // User's channel description/bio
    profile_image_url: string; // URL to user's profile picture
    offline_image_url: string; // URL to custom offline banner image
    view_count: number; // Total view count for the channel
    email?: string; // Email address (only included with certain auth scopes)
    created_at: string; // ISO timestamp of account creation date
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

/**
 * Manages API requests and stream state tracking for Twitch integration
 */
class TwitchAPIManager {
    private oauthToken: string | null = null;
    private expirationTime: number | null = null;
    private onlineStreamers: StreamData[] = [];
    private userData: TwitchUser[] = [];

    /**
     * Creates API request headers with current OAuth token
     * @returns Object containing Authorization and Client-Id headers
     */
    private async getHeaders(): Promise<Record<string, string>> {
        const token = await this.refreshOauthToken();
        return {
            Authorization: `Bearer ${token}`,
            'Client-Id': config.TWITCH_CLIENT_ID,
        };
    }

    /**
     * Refreshes the OAuth token asynchronously if needed
     * @returns A promise that resolves to the current OAuth token
     */
    private async refreshOauthToken(): Promise<string> {
        if (this.oauthToken && this.expirationTime && Date.now() < this.expirationTime) {
            return this.oauthToken;
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
            this.oauthToken = data.access_token;
            this.expirationTime = Date.now() + data.expires_in * 1000;
            logger.debug('Twitch OAuth token refreshed successfully');

            return this.oauthToken;
        } catch (error) {
            logger.error('Error refreshing Twitch OAuth token:', error);
            throw new Error('Failed to refresh Twitch OAuth token');
        }
    }

    /**
     * Gets Twitch user ID from a username/login
     */
    public async getUserIdByLogin(login: string): Promise<string | null> {
        const headers = await this.getHeaders();
        const url = `https://api.twitch.tv/helix/users?login=${login}`;

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

    /**
     * Fetches account information for all monitored streamers
     */
    private async getAccountsData(): Promise<TwitchUser[]> {
        const params = await this.generateDataUrlParams();

        // If we have no streamers to check, return empty array
        if (!params) {
            return [];
        }

        const url = `https://api.twitch.tv/helix/users?${params}`;
        const headers = await this.getHeaders();

        try {
            const response = await axios.get<TwitchUserResponse>(url, { headers });
            const data = response.data.data;
            if (data.length === 0) {
                return [];
            }
            return data;
        } catch (error) {
            logger.error('Error fetching Twitch accounts data:', error);
            throw new Error('Failed to fetch Twitch accounts data');
        }
    }

    /**
     * Fetches stream information for all monitored streamers
     */
    private async getStreamsData(): Promise<StreamData[]> {
        const params = await this.generateOnlineUrlParams();

        // If we have no streamers to check, return empty array
        if (!params) {
            return [];
        }

        const url = `https://api.twitch.tv/helix/streams?${params}`;
        const headers = await this.getHeaders();

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

    /**
     * Gets a list of all unique streamer IDs being monitored
     */
    private async getStreamers(): Promise<string[]> {
        const streamers = await prisma.stream.findMany({
            select: {
                twitchUserId: true,
            },
        });

        return Array.from(new Set(streamers.map((streamer) => streamer.twitchUserId)));
    }

    /**
     * Generates URL parameters for streams API endpoint
     */
    private async generateOnlineUrlParams(): Promise<string> {
        const uniqueStreamers = await this.getStreamers();
        if (uniqueStreamers.length === 0) {
            logger.warn('No streamers found in database for online check');
            return '';
        }
        // Twitch API limits to 100 user_ids per request
        const limitedStreamers = uniqueStreamers.slice(0, 100);
        return limitedStreamers.map((streamer) => `user_id=${streamer}`).join('&');
    }

    /**
     * Generates URL parameters for users API endpoint
     */
    private async generateDataUrlParams(): Promise<string> {
        const uniqueStreamers = await this.getStreamers();
        if (uniqueStreamers.length === 0) {
            logger.warn('No streamers found in database for user data check');
            return '';
        }
        // Twitch API limits to 100 ids per request
        const limitedStreamers = uniqueStreamers.slice(0, 100);
        return limitedStreamers.map((streamer) => `id=${streamer}`).join('&');
    }

    /**
     * Checks for updates in stream status for all monitored streamers
     */
    public async checkStreamsUpdate(): Promise<void> {
        try {
            // Check if we have streamers to monitor before making API calls
            const streamers = await this.getStreamers();
            if (streamers.length === 0) {
                logger.info('No streamers to monitor in database, skipping update check');
                return;
            }

            // Get streams data with validation
            let onlineStreamers: StreamData[] = [];
            try {
                onlineStreamers = await this.getStreamsData();
            } catch (error) {
                logger.error('Failed to get streams data, skipping this update cycle:', error);
            }

            // Get user data with validation
            let userData: TwitchUser[] = [];
            try {
                userData = await this.getAccountsData();
            } catch (error) {
                logger.error('Failed to get account data, continuing with limited functionality:', error);
            }

            // Find streamers with state changes
            const changes = this.detectStreamChanges(onlineStreamers);

            // Process new online streamers
            for (const streamer of changes.newOnline) {
                const userInfo = userData.find((user) => user.id === streamer.user_id);
                if (userInfo) {
                    handleStreamStarted(streamer, userInfo);
                }
            }

            // Process changed streams
            for (const streamer of changes.changed) {
                const userInfo = userData.find((user) => user.id === streamer.user_id);
                if (userInfo) {
                    handleStreamUpdated(streamer, userInfo);
                }
            }

            // Process offline streamers
            for (const streamer of changes.offline) {
                const userInfo = this.userData.find((user) => user.id === streamer.user_id);
                if (userInfo) {
                    handleStreamEnded(userInfo);
                } else {
                    logger.warn(`No user data found for offline streamer: ${streamer.user_name} (${streamer.user_id})`);
                }
            }

            // Update cached data for next check
            this.onlineStreamers = [...onlineStreamers];
            this.userData = [...userData];

            // Only log debug message if there are actual changes
            if (changes.newOnline.length > 0 || changes.changed.length > 0 || changes.offline.length > 0) {
                logger.debug(
                    `Twitch stream check completed: ${changes.newOnline.length} new, ${changes.changed.length} changed, ${changes.offline.length} offline`
                );
            }
        } catch (error) {
            logger.error('Error checking Twitch streams update:', error);
        }
    }

    /**
     * Analyzes streams data to detect changes in status
     */
    private detectStreamChanges(currentStreamers: StreamData[]): {
        newOnline: StreamData[];
        changed: StreamData[];
        offline: StreamData[];
    } {
        // Find streamers who just went online
        const newOnline = currentStreamers.filter((streamer) => {
            return !this.onlineStreamers.some((oldStreamer) => oldStreamer.id === streamer.id);
        });

        // Find streamers who changed their stream details
        const changed = currentStreamers.filter((streamer) => {
            const oldStreamer = this.onlineStreamers.find((old) => old.id === streamer.id);
            if (!oldStreamer) return false;

            return (
                oldStreamer.game_id !== streamer.game_id ||
                oldStreamer.type !== streamer.type ||
                oldStreamer.title !== streamer.title ||
                oldStreamer.viewer_count !== streamer.viewer_count
            );
        });

        // Find streamers who went offline
        const offline = this.onlineStreamers.filter((oldStreamer) => {
            return !currentStreamers.some((streamer) => streamer.id === oldStreamer.id);
        });

        return { newOnline, changed, offline };
    }

    /**
     * Initializes stream data for all monitored streamers
     */
    public async initStreamsUpdate(): Promise<void> {
        try {
            this.onlineStreamers = await this.getStreamsData();
            this.userData = await this.getAccountsData();
            handleInitStreams(this.onlineStreamers, this.userData);
        } catch (error) {
            logger.error('Error initializing Twitch stream update:', error);
        }
    }

    /**
     * Initializes or updates a single streamer's data
     * @param streamerId The Twitch user ID of the streamer to initialize
     * @param forceNotification Whether to force sending a notification even if the streamer is already in the cache
     */
    public async initSingleStreamUpdate(streamerId: string, forceNotification: boolean = false): Promise<void> {
        try {
            const headers = await this.getHeaders();

            // Get user data
            const userUrl = `https://api.twitch.tv/helix/users?id=${streamerId}`;
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

            this.updateUserCache(userInfo);
            const isNewStream = this.updateStreamCache(streamData[0], streamerId);

            // Send notifications based on stream status and force flag
            if (streamData.length > 0 && streamData[0] && (forceNotification || isNewStream)) {
                logger.debug(`Sending stream notification for ${userInfo.display_name}`);
                await handleStreamStarted(streamData[0], userInfo);
            } else if (forceNotification) {
                // Streamer is offline, create an offline embed message
                logger.debug(`Streamer ${userInfo.display_name} is offline, sending offline notification`);
                await handleStreamEnded(userInfo);
            }

            logger.debug(`Single streamer initialization completed for: ${userInfo.display_name} (${streamerId})`);
        } catch (error) {
            logger.error(`Error initializing single streamer (${streamerId}):`, error);
        }
    }

    /**
     * Updates the user cache with new user data
     */
    private updateUserCache(userInfo: TwitchUser): void {
        const userIndex = this.userData.findIndex((user) => user.id === userInfo.id);
        if (userIndex >= 0) {
            this.userData[userIndex] = userInfo;
        } else {
            this.userData.push(userInfo);
        }
    }

    /**
     * Updates the stream cache with new stream data
     * @returns true if this is a newly added stream
     */
    private updateStreamCache(streamInfo: StreamData | undefined, streamerId: string): boolean {
        // If no stream info (offline), just return false
        if (!streamInfo) return false;

        const existingStreamIndex = this.onlineStreamers.findIndex((stream) => stream.user_id === streamerId);

        if (existingStreamIndex >= 0) {
            // Update existing stream data
            this.onlineStreamers[existingStreamIndex] = streamInfo;
            return false;
        } else {
            // Add new stream data to cache
            this.onlineStreamers.push(streamInfo);
            return true;
        }
    }

    /**
     * Removes a streamer from the caches and deletes messages
     * @param streamerId The Twitch user ID of the streamer to remove
     */
    public async removeStreamFromCache(streamerId: string): Promise<void> {
        try {
            // Retrieve streams to delete their messages before they're removed from the database
            const streamsToRemove = await prisma.stream.findMany({
                where: {
                    twitchUserId: streamerId,
                },
            });

            await this.deleteStreamMessages(streamsToRemove);

            // Only remove from cache if no other guild is using this streamer
            const remainingStreamerRefs = await prisma.stream.count({
                where: {
                    twitchUserId: streamerId,
                },
            });

            if (remainingStreamerRefs === 0) {
                this.removeFromCaches(streamerId);
            }
        } catch (error) {
            logger.error(`Error removing streamer from cache (${streamerId}):`, error);
        }
    }

    /**
     * Handles message deletion for removed streams
     */
    private async deleteStreamMessages(streams: {
        uuid: string;
        guildId: string;
        channelId: string;
        roleId: string | null;
        messageId: string | null;
        twitchUserId: string;
    }[]): Promise<void> {
        for (const stream of streams) {
            try {
                logger.debug(
                    `Attempting to delete message for stream ${stream.uuid} in channel ${stream.channelId}, messageId: ${stream.messageId || 'none'}`
                );

                if (stream.messageId) {
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
    }

    /**
     * Removes a streamer from both caches
     */
    private removeFromCaches(streamerId: string): void {
        // Find and remove from onlineStreamers
        const streamIndex = this.onlineStreamers.findIndex((stream) => stream.user_id === streamerId);
        if (streamIndex >= 0) {
            this.onlineStreamers.splice(streamIndex, 1);
        }

        // Find and remove from userData
        const userIndex = this.userData.findIndex((user) => user.id === streamerId);
        if (userIndex >= 0) {
            const userInfo = this.userData[userIndex];
            this.userData.splice(userIndex, 1);

            if (userInfo) {
                logger.debug(`Streamer removed from cache: ${userInfo.display_name} (${streamerId})`);
            } else {
                logger.debug(`Streamer removed from cache: unknown display_name (${streamerId})`);
            }
        }
    }
}

// Create singleton instance
const twitchAPIManager = new TwitchAPIManager();

// Export methods from the manager instance
export const getUserIdByLogin = (login: string) => twitchAPIManager.getUserIdByLogin(login);
export const checkStreamsUpdate = () => twitchAPIManager.checkStreamsUpdate();
export const initStreamsUpdate = () => twitchAPIManager.initStreamsUpdate();
export const initSingleStreamUpdate = (streamerId: string, forceNotification = false) =>
    twitchAPIManager.initSingleStreamUpdate(streamerId, forceNotification);
export const removeStreamFromCache = (streamerId: string) => twitchAPIManager.removeStreamFromCache(streamerId);
