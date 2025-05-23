import { Events, Message, type OmitPartialGroupDMChannel } from 'discord.js';
import type { Event } from '../event';
import { logger } from '../../..';
import { client } from '../../bot';
import { generateWithGoogle } from '../../../utils/intelligence';

export const messageCreateEvent: Event<Events.MessageCreate> = {
    name: Events.MessageCreate,
    once: false,
    execute: async (message) => {
        // Ignore messages from bots
        if (message.author.bot) return;

        if (message.guild) {
            await handleGuildMessage(message);
        } else {
            await message.reply({
                content: 'Je ne peux pas répondre aux messages privés.',
            });
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

    if (message.mentions.has(client.user?.id as string) && !message.mentions.everyone) {
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
}
