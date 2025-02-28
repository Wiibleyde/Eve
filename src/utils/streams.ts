import { WebSocket } from "ws";
import { client, logger } from "..";
import { prisma } from "./database";
import { EmbedBuilder, Message, TextChannel } from "discord.js";

enum WebSocketMessageType {
    MESSAGE = "MESSAGE",
    STREAMS = "STREAMS",
    NEW_STREAM_ONLINE = "NEW_STREAM_ONLINE",
    NEW_STREAM_OFFLINE = "STREAM_OFFLINE",
    ASK_STREAMS = "ASK_STREAMS",
    ADD_STREAM = "ADD_STREAM",
}

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
    id: string,
    login: string,
    display_name: string,
    type: string
    broadcaster_type: string,
    description: string,
    profile_image_url: string,
    offline_image_url: string
    view_count: number
    email?: string
    created_at: string
}

interface WebSocketMessage {
    type: WebSocketMessageType;
    payload: any;
}

// const wsUrl = "ws://localhost:8000";

export class StreamRetrieverConnector {
    private wsUrl: string;
    private ws: WebSocket | null = null;
    private onlineStreams: StreamData[] = [];

    constructor(wsUrl: string) {
        this.wsUrl = wsUrl;
    }

    public start() {
        this.ws = new WebSocket(this.wsUrl);

        this.ws.on("open", () => {
            logger.info("Connected to the stream retriever server.");
        });

        this.ws.on("message", (data) => {
            const message: WebSocketMessage = JSON.parse(data.toString());
            switch (message.type) {
                case WebSocketMessageType.STREAMS:
                    const onlineStreams = message.payload.online;
                    onlineStreams.forEach(async (stream: StreamData) => {
                        this.onlineStreams.push(stream);
                        await onStreamOnline(stream, message.payload.offlineData.find((offlineStream: OfflineStreamData) => offlineStream.login === stream.user_name.toLowerCase()));
                    });
                    const offlineStreams = message.payload.offline;
                    offlineStreams.forEach(async (stream: OfflineStreamData) => {
                        await onStreamOffline({ user_name: stream.login } as StreamData, stream);
                    });
                    break;
                case WebSocketMessageType.NEW_STREAM_ONLINE:
                    this.onlineStreams.push(message.payload);
                    onStreamOnline(message.payload.online, message.payload.offline);
                    break;
                case WebSocketMessageType.NEW_STREAM_OFFLINE:
                    this.onlineStreams = this.onlineStreams.filter((stream) => stream.id !== message.payload.id);
                    onStreamOffline(message.payload.online, message.payload.offline);
                    break;
                default:
                    logger.error("Unknown message type:", JSON.stringify(message.payload));
            }
        });
    }

    private sendMessage(message: WebSocketMessage) {
        if (this.ws) {
            this.ws.send(JSON.stringify(message));
        }
    }

    public askStreams() {
        this.sendMessage({ type: WebSocketMessageType.ASK_STREAMS, payload: null });
    }

    public async addStream(streamer: string) {
        this.sendMessage({ type: WebSocketMessageType.ADD_STREAM, payload: streamer });
    }

    public async removeStream(streamer: string, guildId: string) {
        removeStream(streamer, guildId);
    }
}

function generateEmbed(online: boolean, stream: StreamData, offlineData: OfflineStreamData): EmbedBuilder {
    let embed: EmbedBuilder;
    if(online) {
        embed = new EmbedBuilder()
            .setTitle(offlineData.display_name)
            .setDescription(stream.title)
            .setAuthor({
                name: stream.user_name,
                url: `https://twitch.tv/${stream.user_name}`,
                iconURL: offlineData.profile_image_url,
            })
            .setURL(`https://twitch.tv/${stream.user_name}`)
            .setImage(stream.thumbnail_url.replace("{width}", "1280").replace("{height}", "720"))
            .setThumbnail(`https://static-cdn.jtvnw.net/ttv-boxart/${stream.game_id}.jpg`)
            .addFields(
                {
                    name: "En train de :",
                    value: stream.game_name || "Inconnu",
                    inline: true,
                },
                {
                    name: "Démarré :",
                    value: `<t:${new Date(stream.started_at).getTime() / 1000}:R>`,
                    inline: true,
                }
            )
            .setColor("White")
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client?.user?.displayAvatarURL() })
            .setTimestamp();
    } else {
        embed = new EmbedBuilder()
            .setTitle(offlineData.display_name)
            .setDescription("Le stream est hors-ligne.")
            .setAuthor({
                name: stream.user_name,
                url: `https://twitch.tv/${stream.user_name}`,
                iconURL: offlineData.profile_image_url,
            })
            .setURL(`https://twitch.tv/${stream.user_name}`)
            .setImage(offlineData.offline_image_url)
            .setColor("DarkerGrey")
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client?.user?.displayAvatarURL() })
            .setTimestamp();
    }
    return embed;
}

async function onStreamOnline(stream: StreamData, offlineData: OfflineStreamData) {
    const databaseStreamData = await prisma.stream.findMany({
        where: {
            twitchChannelName: stream.user_name,
        },
    });
    for (const streamData of databaseStreamData) {
        let mention: string = "";
        if (streamData.roleId) {
            mention = `<@&${streamData.roleId}> ${streamData.twitchChannelName} est en ligne !`;
        } else {
            mention = `${streamData.twitchChannelName} est en ligne !`;
        }
        if(!streamData.messageId) {
            const embed = generateEmbed(true, stream, offlineData);
            const channel = await client.channels.fetch(streamData.channelId) as TextChannel;
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
            const channel = await client.channels.fetch(streamData.channelId) as TextChannel;
            if (channel) {
                let message: Message | null = null;
                try {
                    message = await channel.messages.fetch(streamData.messageId as string);
                } catch (error) {
                    logger.error("Error while fetching message:", error);
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
                    if (message.content == mention) {
                        return;
                    }
                    const embed = generateEmbed(true, stream, offlineData);
                    try {
                        await message.delete();
                    } catch (error) {
                        logger.error("Error while deleting message:", error);
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

async function onStreamOffline(stream: StreamData, offlineData: OfflineStreamData) {
    const databaseStreamData = await prisma.stream.findMany({
        where: {
            twitchChannelName: stream.user_name,
        },
    });
    for (const streamData of databaseStreamData) {
        const channel = await client.channels.fetch(streamData.channelId) as TextChannel;
        if (channel) {
            if (streamData.messageId) {
                const message = await channel.messages.fetch(streamData.messageId as string);
                if (message) {
                    if (message.content == `${streamData.twitchChannelName} est en ligne !`) {
                        const embed = generateEmbed(false, stream, offlineData);
                        await message.delete();
                        const newMessage = await channel.send({ embeds: [embed] });
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
            } else {
                const embed = generateEmbed(false, stream, offlineData);
                const message = await channel.send({ embeds: [embed] });
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

async function removeStream(streamer: string, guildId: string) {
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

    const channel = await client.channels.fetch(databaseStreamData.channelId) as TextChannel;
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