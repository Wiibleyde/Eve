import { logger } from '..';
import { client } from '../bot/bot';
import { config } from './config';
import { prisma } from './database';
import { Attachment, Sticker, TextChannel, ThreadChannel } from 'discord.js';

const opennedMp = new Map<string, string>();

export async function initMpThreads() {
    const existingMpThreads = await prisma.mpThreads.findMany({
        select: {
            threadId: true,
            user: {
                select: {
                    userId: true,
                },
            },
        },
    });
    for (const thread of existingMpThreads) {
        if (thread.user?.userId) {
            opennedMp.set(thread.threadId, thread.user.userId);
        }
    }
}

export function handleMessageSend(channelId: string, message: string, stickers: Sticker[], attachments: Attachment[]) {
    if (isNewMessageInMpThread(channelId)) {
        sendMpMessage(channelId, message, stickers, attachments);
    }
}

export function isNewMessageInMpThread(threadId: string) {
    return opennedMp.has(threadId);
}

function sendMpMessage(threadId: string, message: string, stickers: Sticker[], attachments: Attachment[]) {
    if (opennedMp.has(threadId)) {
        const userId = opennedMp.get(threadId);
        if (userId) {
            sendMp(userId, message, stickers, attachments);
        }
    }
}

export async function sendMp(userId: string, message: string, stickers: Sticker[], attachments: Attachment[]) {
    const user = await client.users.fetch(userId);
    try {
        const validStickers = stickers.filter((sticker) => sticker.available);
        await user.send({
            content: message,
            stickers: validStickers.map((sticker) => sticker.id),
            files: attachments.map((attachment) => attachment.url),
        });
    } catch (error) {
        logger.error(error);
    }
}

export async function createMpThread(userId: string): Promise<string> {
    const channel = (await client.channels.fetch(config.MP_CHANNEL)) as TextChannel;
    if (!channel) {
        return '';
    }
    const user = await client.users.fetch(userId);
    if (!user) {
        return '';
    }
    const message = await channel.send({
        content: `Messages privés avec <@${user.id}>`,
    });
    const thread = await message.startThread({
        name: `MP avec ${user.username}`,
        autoArchiveDuration: 60,
    });
    opennedMp.set(thread.id, userId);
    await prisma.mpThreads.create({
        data: {
            threadId: thread.id,
            user: {
                connectOrCreate: {
                    where: {
                        userId,
                    },
                    create: {
                        userId,
                    },
                },
            },
        },
    });
    return thread.id;
}

export async function recieveMessage(userId: string, message: string, stickers: Sticker[], attachments: Attachment[]) {
    let threadId = '';
    for (const [key, value] of opennedMp.entries()) {
        if (value === userId) {
            threadId = key;
            break;
        }
    }
    if (!threadId) {
        threadId = await createMpThread(userId);
    }
    let thread: ThreadChannel;
    try {
        thread = (await client.channels.fetch(threadId)) as ThreadChannel;
    } catch {
        logger.error('Error while fetching the thread, recreating it');
        await prisma.mpThreads.delete({
            where: {
                threadId,
            },
        });
        threadId = await createMpThread(userId);
        thread = (await client.channels.fetch(threadId)) as ThreadChannel;
    }
    const validStickers = stickers.filter((sticker) => sticker.available);
    await thread.send({
        content: `<@${config.OWNER_ID}> ${message}`,
        stickers: validStickers.map((sticker) => sticker.id),
        files: attachments.map((attachment) => attachment.url),
    });
    if (validStickers.length < stickers.length) {
        await thread.send({
            content: "Certains stickers n'ont pas pu être envoyés car ils ne sont pas disponibles.",
        });
    }
}
