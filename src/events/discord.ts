import { Events, MessageType } from 'discord.js';
import { client, logger } from '..';
import { maintenance } from '@/interactions/commands/dev/maintenance';
import { hasPermission } from '@/utils/permissionTester';
import { errorEmbed } from '@/utils/embeds';
import { commands, devCommands } from '@/interactions/commands';
import { handleMessageSend, isNewMessageInMpThread, recieveMessage } from '@/utils/mpManager';
import { config } from '@/config';
import { isMessageQuizQuestion } from '@/interactions/commands/fun/quiz/quiz';
import { generateWithGoogle } from '@/utils/intelligence';
import { detectFeur, generateResponse } from '@/utils/messageManager';
import { contextMessageMenus, contextUserMenus } from '@/interactions/contextMenus';
import { modals } from '@/interactions/modals';
import { buttons } from '@/interactions/buttons';
import { selectMenus } from '@/interactions/selectMenus';

client.on(Events.InteractionCreate, async (interaction) => {
    if (interaction.isContextMenuCommand()) {
        if (interaction.isMessageContextMenuCommand()) {
            const commandName = interaction.commandName as keyof typeof contextMessageMenus;
            contextMessageMenus[commandName].execute(interaction);
        } else if (interaction.isUserContextMenuCommand()) {
            const commandName = interaction.commandName as keyof typeof contextUserMenus;
            contextUserMenus[commandName].execute(interaction);
        }
        logger.info(
            `Commande contextuelle </${interaction.commandName}:${interaction.commandId}> par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}>`
        );
    } else if (interaction.isCommand()) {
        if (maintenance) {
            if (!(await hasPermission(interaction, [], false))) {
                await interaction.reply({
                    embeds: [
                        errorEmbed(interaction, new Error('Le bot est en maintenance, veuillez réessayer plus tard.')),
                    ],
                    ephemeral: true,
                });
                return;
            }
        }
        const { commandName } = interaction;
        if (commands[commandName as keyof typeof commands]) {
            commands[commandName as keyof typeof commands].execute(interaction);
        }
        if (devCommands[commandName as keyof typeof devCommands]) {
            devCommands[commandName as keyof typeof devCommands].execute(interaction);
        }
        logger.info(
            `Commande </${commandName}:${interaction.commandId}> par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}>`
        );
    } else if (interaction.isModalSubmit()) {
        const customId = interaction.customId.split('--')[0];
        if (modals[customId as keyof typeof modals]) {
            modals[customId as keyof typeof modals](interaction);
        }
        logger.info(
            `Modal soumis par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId})`
        );
    } else if (interaction.isButton()) {
        const customId = interaction.customId.split('--')[0];
        if (buttons[customId as keyof typeof buttons]) {
            buttons[customId as keyof typeof buttons](interaction);
        }
        logger.info(
            `Bouton cliqué par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId})`
        );
    } else if (interaction.isStringSelectMenu()) {
        const customId = interaction.customId.split('--')[0];
        logger.info(
            `Menu déroulant par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId})`
        );
        if (selectMenus[customId as keyof typeof selectMenus]) {
            selectMenus[customId as keyof typeof selectMenus](interaction);
        }
    } else {
        logger.error(
            `Interaction inconnue par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}>`
        );
    }
});

client.on(Events.MessageCreate, async (message) => {
    if (message.author.bot) return;

    const guildId = message.guild?.id;
    if (!guildId) {
        const messageStickers = Array.from(message.stickers.values());
        const messageAttachments = Array.from(message.attachments.values());
        recieveMessage(message.author.id, message.content, messageStickers, messageAttachments);
        return;
    }
    if (
        guildId === config.EVE_HOME_GUILD &&
        message.author.id != client.user?.id &&
        isNewMessageInMpThread(message.channel.id)
    ) {
        const messageStickers = Array.from(message.stickers.values());
        const messageAttachments = Array.from(message.attachments.values());
        handleMessageSend(message.channel.id, message.content, messageStickers, messageAttachments);
        return;
    }

    const channelId = message?.channel?.id;
    if (!channelId) {
        logger.error(`Channel non trouvé pour le message de <@${message.author.id}> dans le serveur ${guildId}`);
        return;
    }

    if (message.mentions.has(client.user?.id as string) && !message.mentions.everyone) {
        if (message.type === MessageType.Reply) {
            if (isMessageQuizQuestion(message.reference?.messageId as string)) {
                return;
            }
            // const contentOfReply = message.reference?.messageId ? await message.channel.messages.fetch(message.reference.messageId).then(msg => msg.content) : ''
            // message.content = contentOfReply + message.content
        }
        message.channel.sendTyping();
        const aiReponse = await generateWithGoogle(
            channelId,
            message.content.replace(`<@${client.user?.id}> `, ''),
            message.author.id
        )
            .catch((error) => {
                return (
                    'Je ne suis pas en mesure de répondre à cette question pour le moment. ||(' +
                    error +
                    ')|| (Conversation réinitialisée)'
                );
            })
            .then((response) => {
                return response;
            });

        if (aiReponse) {
            await message.channel.send(`${aiReponse}`);
            logger.info(`Réponse de l'IA à <@${message.author.id}> dans <#${channelId}> : ${aiReponse}`);
        }
    }

    if (detectFeur(message.content)) {
        message.channel.send(generateResponse());
    }
});
