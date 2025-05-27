import { logger } from '..';
import { client } from '../bot/bot';
import { config } from './core/config';
import { prisma } from './core/database';
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
    const userId = opennedMp.get(threadId);
    if (userId) {
        sendMp(userId, message, stickers, attachments);
    }
}

/**
 * Sends a direct message to a user
 */
export async function sendMp(
    userId: string,
    message: string,
    stickers: Sticker[] = [],
    attachments: Attachment[] = []
) {
    try {
        const user = await client.users.fetch(userId);
        const validStickers = getValidStickers(stickers);

        await user.send({
            content: message,
            stickers: validStickers.map((sticker) => sticker.id),
            files: attachments.map((attachment) => attachment.url),
        });
    } catch (error) {
        logger.error(`Error sending MP to user ${userId}:`, error);
    }
}

/**
 * Creates a new thread for private messages with a user
 */
export async function createMpThread(userId: string): Promise<string> {
    try {
        const channel = (await client.channels.fetch(config.MP_CHANNEL)) as TextChannel;
        if (!channel) {
            logger.error('MP channel not found');
            return '';
        }

        const user = await client.users.fetch(userId);
        if (!user) {
            logger.error(`User ${userId} not found`);
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
                        where: { userId },
                        create: { userId },
                    },
                },
            },
        });

        return thread.id;
    } catch (error) {
        logger.error(`Error creating MP thread for user ${userId}:`, error);
        return '';
    }
}

/**
 * Finds or creates a thread for a user
 */
async function findOrCreateThreadForUser(userId: string): Promise<ThreadChannel | null> {
    // Find existing thread
    let threadId = findThreadIdForUser(userId);

    // Create new thread if none exists
    if (!threadId) {
        threadId = await createMpThread(userId);
        if (!threadId) return null;
    }

    // Fetch the thread
    try {
        return (await client.channels.fetch(threadId)) as ThreadChannel;
    } catch (error) {
        logger.error(`Error fetching thread ${threadId}, recreating it:`, error);

        // Cleanup failed thread
        await prisma.mpThreads.delete({ where: { threadId } });
        opennedMp.delete(threadId);

        // Create new thread
        threadId = await createMpThread(userId);
        if (!threadId) return null;

        return (await client.channels.fetch(threadId)) as ThreadChannel;
    }
}

/**
 * Finds the thread ID for a user
 */
function findThreadIdForUser(userId: string): string {
    for (const [threadId, id] of opennedMp.entries()) {
        if (id === userId) {
            return threadId;
        }
    }
    return '';
}

/**
 * Filters out stickers that aren't available
 */
function getValidStickers(stickers: Sticker[]): Sticker[] {
    return stickers.filter((sticker) => sticker.available);
}

/**
 * Handles receiving a message from a user
 */
export async function recieveMessage(
    userId: string,
    message: string,
    stickers: Sticker[] = [],
    attachments: Attachment[] = []
) {
    const thread = await findOrCreateThreadForUser(userId);
    if (!thread) {
        logger.error(`Failed to find or create thread for user ${userId}`);
        return;
    }

    const validStickers = getValidStickers(stickers);

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
