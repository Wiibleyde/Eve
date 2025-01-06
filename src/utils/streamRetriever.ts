import { WebSocket } from "ws";
import { client, logger } from "..";
import { prisma } from "./database";
import { CategoryChannel, EmbedBuilder, GuildBasedChannel, Message, TextChannel } from "discord.js";

enum WebSocketMessageType {
    MESSAGE = "MESSAGE",
    STREAMS = "STREAMS",
    NEW_STREAM_ONLINE = "NEW_STREAM_ONLINE",
    NEW_STREAM_OFFLINE = "STREAM_OFFLINE",
    ASK_STREAMS = "ASK_STREAMS",
    ASK_STREAM = "ASK_STREAM",
    ADD_STREAM = "ADD_STREAM",
}

interface WebSocketMessage {
    type: WebSocketMessageType;
    payload: any;
}

interface StreamData {
    id: string;
    user_id: string;
    user_login: string;
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

export class StreamRetriever {
    private ws!: WebSocket
    private retryTimeout: number = 10000; // Retry after 10 seconds
    private reconnecting: boolean = false;

    constructor(private url: string) {
        this.ws = new WebSocket(this.url)
        this.init()
    }

    public init() {
        this.ws.on("open", () => this.onOpen(this.ws))
        this.ws.on("message", (message) => this.onMessage(this.ws, JSON.parse(message.toString())))
        this.ws.on("close", () => this.onClose(this.ws))
        this.ws.on("error", (error) => this.onError(error))
    }

    private onOpen(ws: WebSocket) {
        logger.info(`Connected to ${this.url}`)
    }

    private onMessage(ws: WebSocket, message: WebSocketMessage) {
        switch (message.type) {
            case WebSocketMessageType.MESSAGE:
                logger.debug(`Received message: ${message.payload}`)
                break
            case WebSocketMessageType.STREAMS:
                const streams: StreamData[] = message.payload
                this.checkStreams(streams)
                break
            case WebSocketMessageType.NEW_STREAM_ONLINE:
                const streamDataOnline: StreamData = message.payload
                this.handleStreamOnline(streamDataOnline)
                break
            case WebSocketMessageType.NEW_STREAM_OFFLINE:
                const streamDataOffline: StreamData = message.payload
                this.handleStreamOffline(streamDataOffline)
                break
            default:
                logger.warn(`Unknown message type: ${message.type}`)
        }
    }

    private onClose(ws: WebSocket) {
        logger.debug(`Disconnected from ${this.url}`);
        if (!this.reconnecting) {
            this.reconnect();
        }
    }

    private onError(error: any) {
        if (error.code === 'ECONNREFUSED') {
            logger.error(`Connection refused to ${this.url}. Retrying in ${this.retryTimeout / 1000} seconds...`);
            if (!this.reconnecting) {
                this.reconnect();
            }
        } else {
            logger.error(`WebSocket error: ${error.message}`);
        }
    }

    private reconnect() {
        this.reconnecting = true;
        setTimeout(() => {
            logger.debug(`Reconnecting to ${this.url}`);
            if (this.ws.readyState !== WebSocket.CLOSED) {
                this.ws.terminate(); // Abort the old connection
            }
            this.ws = new WebSocket(this.url);
            this.init();
            this.reconnecting = false;
        }, this.retryTimeout);
    }

    private send(message: WebSocketMessage, ws: WebSocket): void {
        ws.send(JSON.stringify(message));
    }

    private async handleStreamOnline(streamData: StreamData) {
        const databaseStreamDatas = await prisma.streams.findMany({
            where: {
                channelName: streamData.user_name
            }
        });

        if (!databaseStreamDatas || databaseStreamDatas.length === 0) {
            logger.warn(`Received stream online event for unknown stream: ${streamData.user_name}`);
            return;
        }

        const { messages, channels, newMessageToSend } = await this.prepareMessagesAndChannels(databaseStreamDatas, streamData, "online");

        await this.updateMessages(messages, channels, databaseStreamDatas, streamData, "online");
        await this.sendNewMessages(newMessageToSend, channels, databaseStreamDatas, streamData, "online");
    }

    private async handleStreamOffline(streamData: StreamData) {
        const databaseStreamDatas = await prisma.streams.findMany({
            where: {
                channelName: streamData.user_name
            }
        });

        if (!databaseStreamDatas || databaseStreamDatas.length === 0) {
            logger.warn(`Received stream offline event for unknown stream: ${streamData.user_name}`);
            return;
        }

        const { messages, channels, newMessageToSend } = await this.prepareMessagesAndChannels(databaseStreamDatas, streamData, "offline");

        await this.updateMessages(messages, channels, databaseStreamDatas, streamData, "offline");
        await this.sendNewMessages(newMessageToSend, channels, databaseStreamDatas, streamData, "offline");
    }

    private async prepareMessagesAndChannels(databaseStreamDatas: any[], streamData: StreamData, type: "online" | "offline") {
        const messages = new Map<string, Message>();
        const channels = new Map<string, GuildBasedChannel>();

        for (const databaseStreamData of databaseStreamDatas) {
            if (databaseStreamData.messageSentId) {
                const guild = await client.guilds.fetch(databaseStreamData.guildId);
                const channel = await guild.channels.fetch(databaseStreamData.discordChannelId);
                if (channel && channel.isTextBased()) {
                    channels.set(databaseStreamData.guildId, channel);
                    const message = await (channel as TextChannel).messages.fetch(databaseStreamData.messageSentId);
                    if (message) {
                        // Check if the message is already correct
                        const isCorrectMessage = type === "online"
                            ? message.embeds[0]?.title === streamData.title && message.embeds[0]?.url === `https://twitch.tv/${streamData.user_name}`
                            : message.embeds[0]?.title === "Stream hors ligne" && message.embeds[0]?.url === `https://twitch.tv/${streamData.user_name}`;
                        if (isCorrectMessage) {
                            continue;
                        }
                        messages.set(databaseStreamData.guildId, message);
                    }
                }
            }
        }

        const newMessageToSend = new Map<string, string>();
        // If one channel has no message, remove it from the list
        if (channels.size !== messages.size) {
            for (const [guildId, channel] of channels) {
                if (!messages.has(guildId)) {
                    channels.delete(guildId);
                    newMessageToSend.set(guildId, streamData.user_name);
                }
            }
        }

        return { messages, channels, newMessageToSend };
    }

    private async updateMessages(messages: Map<string, Message>, channels: Map<string, GuildBasedChannel>, databaseStreamDatas: any[], streamData: StreamData, type: "online" | "offline") {
        for (const [guildId, message] of messages) {
            // Remove older message and send a new one with the new stream data
            await message.delete();
            const channel = channels.get(guildId);
            const databaseStreamData = databaseStreamDatas.find((streamData) => streamData.guildId === guildId);
            if (channel) {
                const content = databaseStreamData?.roleId ? `<@${databaseStreamData.roleId}>` : '';
                const embed = this.createEmbed(streamData, type);

                if (channel instanceof TextChannel) {
                    const newMessage = await channel.send({ content, embeds: [embed] });
                    if (databaseStreamData) {
                        await prisma.streams.update({
                            where: { id: databaseStreamData.id },
                            data: { messageSentId: newMessage.id }
                        });
                    }
                }
            }
        }
    }

    private async sendNewMessages(newMessageToSend: Map<string, string>, channels: Map<string, GuildBasedChannel>, databaseStreamDatas: any[], streamData: StreamData, type: "online" | "offline") {
        for (const [guildId, channelName] of newMessageToSend) {
            const channel = channels.get(guildId);
            if (channel) {
                const databaseStreamData = databaseStreamDatas.find((streamData) => streamData.guildId === guildId);
                const content = databaseStreamData?.roleId ? `<@${databaseStreamData.roleId}>` : '';
                const embed = this.createEmbed(streamData, type);

                if (channel instanceof TextChannel) {
                    const newMessage = await channel.send({ content, embeds: [embed] });
                    if (databaseStreamData) {
                        await prisma.streams.update({
                            where: { id: databaseStreamData.id },
                            data: { messageSentId: newMessage.id }
                        });
                    }
                }
            }
        }
    }

    private createEmbed(streamData: StreamData, type: "online" | "offline") {
        logger.debug(`Creating embed for stream ${streamData.user_name} with type ${type}`)
        const title = type === "online" ? streamData.title : "Stream hors ligne";
        const description = type === "online"
            ? `**${streamData.user_name}** est en live sur [Twitch](https://twitch.tv/${streamData.user_name})`
            : `**${streamData.user_name}** est hors ligne sur [Twitch](https://twitch.tv/${streamData.user_name})`;

        return new EmbedBuilder()
            .setTitle(title)
            .setURL(`https://twitch.tv/${streamData.user_name}`)
            .setDescription(description)
            .setImage(streamData.thumbnail_url)
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user ? client.user.displayAvatarURL() : '' })
            .setTimestamp();
    }

    private async checkStreams(streams: StreamData[]) {
        logger.debug(`Checking streams...`, JSON.stringify(streams))
        const databaseStreamDatas = await prisma.streams.findMany()
        // Check if the message was sent and if correct information (offline message for offline stream and online message for online stream)
        for (const databaseStreamData of databaseStreamDatas) {
            logger.debug(`Checking stream ${databaseStreamData.channelName}`)
            const streamData = streams.find((stream) => stream.user_login === databaseStreamData.channelName)
            logger.debug(`Stream data: ${JSON.stringify(streamData)}`)
            if (streamData) {
                logger.debug(`Stream ${streamData.user_name} is online: ${streamData.type === "live"}`)
                if (streamData.type === "live") {
                    if (!databaseStreamData.messageSentId) {
                        logger.debug(`Stream ${streamData.user_name} is online but no message was sent. Sending a new message...`)
                        await this.handleStreamOnline(streamData)
                    }
                } else {
                    await this.handleStreamOffline(streamData)
                }
            }
        }
    }

    public addStream(channelName: string) {
        const message: WebSocketMessage = {
            type: WebSocketMessageType.ADD_STREAM,
            payload: channelName
        }
        this.send(message, this.ws)
    }
}
