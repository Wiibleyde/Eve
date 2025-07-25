import { Events, Message, type OmitPartialGroupDMChannel } from 'discord.js';
import type { Event } from '../event';
import { logger } from '../../..';
import { client } from '../../bot';
import { generateWithGoogle, iaDisabledServers } from '../../../utils/intelligence';
import { isMaintenanceMode } from '../../../utils/core/maintenance';
import { warningEmbedGenerator } from '../../utils/embeds';
import { config } from '../../../utils/core/config';
import { handleMessageSend, isNewMessageInMpThread, recieveMessage } from '../../../utils/mpManager';
import { detectPattern, generateResponse, jokeIgnoredServers } from '../../../utils/messageManager';

export const messageCreateEvent: Event<Events.MessageCreate> = {
    name: Events.MessageCreate,
    once: false,
    execute: async (message) => {
        // Ignore messages from bots
        if (message.author.bot) return;

        if (isMaintenanceMode() && message.author.id !== config.OWNER_ID) {
            await message.reply({
                embeds: [
                    warningEmbedGenerator('Le bot est actuellement en mode maintenance. Veuillez réessayer plus tard.'),
                ],
            });
            return;
        }

        if (message.guild) {
            await handleGuildMessage(message);
        } else {
            await handleDirectMessage(message);
        }
    },
};

async function handleGuildMessage(message: OmitPartialGroupDMChannel<Message<boolean>>) {
    const startTime = Date.now();
    const guildId = message.guild?.id;
    const channelId = message.channel?.id;

    if (!guildId || !channelId) {
        logger.error(`Guild ou channel non trouvé pour le message de <@${message.author.id}>`);
        return;
    }

    if (
        guildId === config.EVE_HOME_GUILD &&
        message.author.id !== client.user?.id &&
        isNewMessageInMpThread(channelId)
    ) {
        const messageStickers = Array.from(message.stickers.values());
        const messageAttachments = Array.from(message.attachments.values());
        handleMessageSend(channelId, message.content, messageStickers, messageAttachments);
        return;
    }

    if (message.mentions.has(client.user?.id as string) && !message.mentions.everyone) {
        const guildIdStr = String(guildId);
        if (iaDisabledServers.includes(guildIdStr)) {
            logger.info(
                `IA désactivée pour le serveur ${guildIdStr}. Message ignoré de <@${message.author.id}> : "${message.content}"`
            );
            return;
        }
        message.channel.sendTyping();
        const aiResponse = await generateWithGoogle(channelId, message.content, message.author.id).catch((error) => {
            logger.error(`Erreur lors de la génération de réponse IA: ${error}`);
            return 'Je ne suis pas en mesure de répondre à cette question pour le moment. (Conversation réinitialisée)';
        });

        if (aiResponse) {
            await message.channel.send(aiResponse);
            logger.info(
                `[${Date.now() - startTime}ms] Réponse de l'IA envoyée dans <#${channelId}> par <@${message.author.id}> : "${message.content}"  : ${aiResponse}`
            );
        }
    }

    const pattern = detectPattern(message.content);
    if (pattern) {
        const guildIdStr = String(guildId);
        if (jokeIgnoredServers.includes(guildIdStr)) {
            logger.info(
                `Message ignoré dans le serveur ${guildIdStr} de <@${message.author.id}> : "${message.content}"`
            );
            return;
        }
        const response = generateResponse(pattern);
        if (response) {
            await message.channel.send(response);
            logger.info(
                `[${Date.now() - startTime}ms] Réponse générée pour le motif "${pattern.name}" dans <#${channelId}> par <@${message.author.id}> : "${message.content}"  : ${response}`
            );
        }
    }
}

async function handleDirectMessage(message: OmitPartialGroupDMChannel<Message<boolean>>) {
    const startTime = Date.now();
    const userId = message.author.id;

    const messageStickers = Array.from(message.stickers.values());
    const messageAttachments = Array.from(message.attachments.values());

    await recieveMessage(userId, message.content, messageStickers, messageAttachments);

    logger.info(
        `[${Date.now() - startTime}ms] Message privé reçu de <@${userId}> : "${message.content}" avec ${messageStickers.length} stickers et ${messageAttachments.length} pièces jointes`
    );
}
