import { config } from '@/config';
import { logger } from '..';
import { prisma } from './database';
import { onStreamOffline, onStreamOnline } from './streams';

const CLIENT_ID = config.TWITCH_CLIENT_ID;
const CLIENT_SECRET = config.TWITCH_CLIENT_SECRET;

export class TwitchService {
    private OAUTH_TOKEN = '';
    private tokenExpiryTime = 0;
    private refreshInterval = 6000; // 6 seconds
    public onlineStreamers: StreamData[] = [];
    public offlineStreamDatas: OfflineStreamData[] = [];

    constructor() {
        this.startAutoRefresh();
    }

    private async getOAuthToken(): Promise<OAuthResponse | null> {
        const url = 'https://id.twitch.tv/oauth2/token';
        const params = new URLSearchParams({
            client_id: CLIENT_ID,
            client_secret: CLIENT_SECRET,
            grant_type: 'client_credentials',
        });

        try {
            const response = await fetch(url, {
                method: 'POST',
                body: params,
            });
            const data: OAuthResponse = await response.json();
            this.OAUTH_TOKEN = data.access_token;
            this.tokenExpiryTime = Date.now() + data.expires_in * 1000; // Set token expiry time
            return data;
        } catch (error) {
            logger.error('Erreur lors de la récupération du token OAuth :', error);
            return null;
        }
    }

    private async areStreamersLive(): Promise<StreamData[]> {
        // Refresh the token if it has expired
        if (Date.now() >= this.tokenExpiryTime) {
            await this.getOAuthToken();
        }
        const theUrl = `https://api.twitch.tv/helix/streams?${await this.generateUrlParams()}`;
        const headers = {
            'Client-Id': CLIENT_ID,
            Authorization: `Bearer ${this.OAUTH_TOKEN}`,
        };

        try {
            const response = await fetch(theUrl, { headers });
            const data: StreamResponse = await response.json();
            const liveStreamers = data.data;
            return liveStreamers;
        } catch (error) {
            logger.error('Erreur lors de la récupération des streamers en direct :', error);
            return [];
        }
    }

    private async getOfflineStreamsData(): Promise<OfflineStreamData[]> {
        if (Date.now() >= this.tokenExpiryTime) {
            await this.getOAuthToken();
        }
        const url = `https://api.twitch.tv/helix/users?${await this.generateOfflineUrlParams()}`;
        const headers = {
            'Client-Id': CLIENT_ID,
            Authorization: `Bearer ${this.OAUTH_TOKEN}`,
        };

        try {
            const response = await fetch(url, { headers });
            const data: OfflineStreamResponse = await response.json();
            const offlineStreamers = data.data;
            return offlineStreamers;
        } catch (error) {
            logger.error('Erreur lors de la récupération des streamers hors ligne :', error);
            return [];
        }
    }

    public getOfflineStreamData(login: string): OfflineStreamData | undefined {
        if (!this.offlineStreamDatas) {
            logger.error('Aucune donnée hors ligne trouvée');
            return undefined;
        }
        return this.offlineStreamDatas.find((streamer) => streamer.login.toLowerCase() === login.toLowerCase());
    }

    public getStreamData(login: string): StreamData | undefined {
        if (!this.onlineStreamers) {
            logger.error('Aucune donnée en ligne trouvée');
            return undefined;
        }
        return this.onlineStreamers.find((streamer) => streamer.user_name.toLowerCase() === login.toLowerCase());
    }

    private async generateUrlParams(): Promise<string> {
        const streamers = await this.getStreamers();
        const uniqueStreamers = Array.from(new Set(streamers));
        const limitedStreamers = uniqueStreamers.slice(0, 100);
        return limitedStreamers.map((streamer) => `user_login=${streamer}`).join('&');
    }

    private async generateOfflineUrlParams(): Promise<string> {
        const streamers = await this.getStreamers();
        const uniqueStreamers = Array.from(new Set(streamers));
        const limitedStreamers = uniqueStreamers.slice(0, 100);
        return limitedStreamers.map((streamer) => `login=${streamer}`).join('&');
    }

    private async startAutoRefresh(): Promise<void> {
        await this.getOAuthToken();
        this.onlineStreamers = await this.areStreamersLive();
        this.offlineStreamDatas = await this.getOfflineStreamsData();

        if (this.offlineStreamDatas) {
            for (const streamer of this.onlineStreamers) {
                const offlineData = this.getOfflineStreamData(streamer.user_name);
                if (offlineData) {
                    onStreamOnline(streamer, offlineData);
                } else {
                    logger.error(`Aucune donnée hors ligne trouvée pour ${streamer.user_name}`);
                }
            }
            for (const offlineStreamer of this.offlineStreamDatas) {
                if (
                    this.onlineStreamers.some(
                        (streamer) => streamer.user_name.toLowerCase() === offlineStreamer.login.toLowerCase()
                    )
                ) {
                    continue;
                }
                onStreamOffline(offlineStreamer);
            }
        }

        setInterval(async () => {
            this.onlineStreamers = await this.areStreamersLive();
            this.offlineStreamDatas = await this.getOfflineStreamsData();

            if (this.offlineStreamDatas) {
                for (const streamer of this.onlineStreamers) {
                    const offlineData = this.getOfflineStreamData(streamer.user_name);
                    if (offlineData) {
                        onStreamOnline(streamer, offlineData);
                    } else {
                        logger.error(`Aucune donnée hors ligne trouvée pour ${streamer.user_name}`);
                    }
                }
                for (const offlineStreamer of this.offlineStreamDatas) {
                    if (
                        this.onlineStreamers.some(
                            (streamer) => streamer.user_name.toLowerCase() === offlineStreamer.login.toLowerCase()
                        )
                    ) {
                        continue;
                    }
                    onStreamOffline(offlineStreamer);
                }
            }
        }, this.refreshInterval);
    }

    private async getStreamers(): Promise<string[]> {
        const streamers = await prisma.stream.findMany({
            select: {
                twitchChannelName: true,
            },
        });
        return streamers.map((streamer) => streamer.twitchChannelName);
    }
}

interface OAuthResponse {
    access_token: string;
    expires_in: number;
    token_type: string;
}

interface StreamPagination {
    cursor: string;
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

export interface OfflineStreamData {
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

export interface StreamResponse {
    data: StreamData[];
    pagination: StreamPagination;
}

export interface OfflineStreamResponse {
    data: OfflineStreamData[];
}
