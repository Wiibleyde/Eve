import {
    ApplicationCommandOptionType,
    ButtonInteraction,
    CacheType,
    ChatInputCommandInteraction,
    Events,
    Message,
    MessageContextMenuCommandInteraction,
    MessageFlags,
    MessageType,
    ModalSubmitInteraction,
    OmitPartialGroupDMChannel,
    StringSelectMenuInteraction,
    UserContextMenuCommandInteraction,
} from 'discord.js';
import { client, logger } from '..';
import { maintenance } from '@/interactions/commands/dev/maintenance';
import { hasPermission } from '@/utils/permissionTester';
import { errorEmbed } from '@/utils/embeds';
import { commands, devCommands } from '@/interactions/commands';
import { handleMessageSend, isNewMessageInMpThread, recieveMessage } from '@/utils/mpManager';
import { config } from '@/config';
import { isMessageQuizQuestion } from '@/interactions/commands/fun/quiz/quiz';
import { generateWithGoogle } from '@/utils/intelligence';
import { detectFeur, generateFeurResponse } from '@/utils/messageManager';
import { contextMessageMenus, contextUserMenus } from '@/interactions/contextMenus';
import { modals } from '@/interactions/modals';
import { buttons } from '@/interactions/buttons';
import { selectMenus } from '@/interactions/selectMenus';

async function handleContextMenu(
    interaction: MessageContextMenuCommandInteraction<CacheType> | UserContextMenuCommandInteraction<CacheType>
) {
    try {
        if (interaction.isMessageContextMenuCommand()) {
            const commandName = interaction.commandName as keyof typeof contextMessageMenus;
            await contextMessageMenus[commandName]?.execute(interaction);
        } else if (interaction.isUserContextMenuCommand()) {
            const commandName = interaction.commandName as keyof typeof contextUserMenus;
            await contextUserMenus[commandName]?.execute(interaction);
        }
        logger.info(
            `Commande contextuelle </${interaction.commandName}:${interaction.commandId}> par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}>`
        );
    } catch (error) {
        logger.error(`Erreur lors de l'exécution d'une commande contextuelle: ${error}`);
        if (interaction.replied) {
            await interaction.followUp({
                embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}

async function handleCommand(interaction: ChatInputCommandInteraction) {
    try {
        if (maintenance && !(await hasPermission(interaction, [], false))) {
            await interaction.reply({
                embeds: [
                    errorEmbed(interaction, new Error('Le bot est en maintenance, veuillez réessayer plus tard.')),
                ],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }

        const { commandName } = interaction;
        await commands[commandName as keyof typeof commands]?.execute(interaction);
        await devCommands[commandName as keyof typeof devCommands]?.execute(interaction);

        const cmdArgs = interaction.options.data.map((option) => {
            if (option.type === ApplicationCommandOptionType.Subcommand) {
                return `${option.name} ${option.options?.map((opt) => `${opt.name}:${opt.value}`).join(' ')}`;
            }
            return `${option.name}:${option.value}`;
        });

        logger.info(
            `Commande </${commandName}:${interaction.commandId}> par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${cmdArgs.join(' ')})`
        );
    } catch (error) {
        logger.error(`Erreur lors de l'exécution d'une commande: ${error}`);
        if (interaction.replied) {
            await interaction.followUp({
                embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}

async function handleModal(interaction: ModalSubmitInteraction) {
    try {
        const customId = interaction.customId.split('--')[0];
        await modals[customId as keyof typeof modals]?.(interaction);
        const modalArgs = interaction.fields.fields.map((field) => {
            return `${field.customId}:${field.value}`;
        });
        logger.info(
            `Modal soumis par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId}) (${modalArgs.join(' ')})`
        );
    } catch (error) {
        logger.error(`Erreur lors de la soumission d'un modal: ${error}`);
        if (interaction.replied) {
            await interaction.followUp({
                embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}

async function handleButton(interaction: ButtonInteraction) {
    try {
        const customId = interaction.customId.split('--')[0];
        await buttons[customId as keyof typeof buttons]?.(interaction);
        logger.info(
            `Bouton cliqué par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId})`
        );
    } catch (error) {
        logger.error(`Erreur lors du clic sur un bouton: ${error}`);
        if (interaction.replied) {
            await interaction.followUp({
                embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}

async function handleSelectMenu(interaction: StringSelectMenuInteraction<CacheType>) {
    try {
        const customId = interaction.customId.split('--')[0];
        await selectMenus[customId as keyof typeof selectMenus]?.(interaction);
        const selectedOptions = interaction.values.map((value) => {
            return value;
        });
        logger.info(
            `Menu déroulant par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}> (${interaction.customId}) (${selectedOptions.join(' ')})`
        );
    } catch (error) {
        logger.error(`Erreur lors de l'utilisation d'un menu déroulant: ${error}`);
        if (interaction.replied) {
            await interaction.followUp({
                embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aïe, une erreur est survenue ||' + error + '||'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}

client.on(Events.InteractionCreate, async (interaction) => {
    if (maintenance && !(await hasPermission(interaction, [], false))) {
        if (interaction.isRepliable()) {
            await interaction.reply({
                embeds: [
                    errorEmbed(interaction, new Error('Le bot est en maintenance, veuillez réessayer plus tard.')),
                ],
                flags: [MessageFlags.Ephemeral],
            });
        }
        return;
    }

    if (interaction.isContextMenuCommand()) {
        await handleContextMenu(interaction);
    } else if (interaction.isChatInputCommand()) {
        await handleCommand(interaction);
    } else if (interaction.isModalSubmit()) {
        await handleModal(interaction);
    } else if (interaction.isButton()) {
        await handleButton(interaction);
    } else if (interaction.isStringSelectMenu()) {
        await handleSelectMenu(interaction);
    } else {
        logger.error(
            `Interaction inconnue par <@${interaction.user.id}> (${interaction.user.username}) dans <#${interaction.channelId}>`
        );
    }
});

async function handleDirectMessage(message: OmitPartialGroupDMChannel<Message<boolean>>) {
    const messageStickers = Array.from(message.stickers.values());
    const messageAttachments = Array.from(message.attachments.values());
    recieveMessage(message.author.id, message.content, messageStickers, messageAttachments);
}

async function handleGuildMessage(message: OmitPartialGroupDMChannel<Message<boolean>>) {
    const guildId = message.guild?.id;
    const channelId = message.channel?.id;

    if (!guildId || !channelId) {
        logger.error(`Guild ou channel non trouvé pour le message de <@${message.author.id}>`);
        return;
    }

    if (maintenance) {
        logger.info(`Message ignoré pendant la maintenance de <@${message.author.id}> dans <#${channelId}>`);
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
        if (message.type === MessageType.Reply && isMessageQuizQuestion(message.reference?.messageId as string)) {
            return;
        }

        message.channel.sendTyping();
        const aiResponse = await generateWithGoogle(
            channelId,
            message.content,
            message.author.id
        ).catch((error) => {
            logger.error(`Erreur lors de la génération de réponse IA: ${error}`);
            return 'Je ne suis pas en mesure de répondre à cette question pour le moment. (Conversation réinitialisée)';
        });

        if (aiResponse) {
            await message.channel.send(aiResponse);
            logger.info(`Réponse de l'IA à <@${message.author.id}> : "${message.content}"  dans <#${channelId}> : ${aiResponse}`);
        }
    }

    if (detectFeur(message.content)) {
        message.channel.send(generateFeurResponse());
    }
}

client.on(Events.MessageCreate, async (message) => {
    if (message.author.bot) return;

    if (!message.guild) {
        await handleDirectMessage(message);
    } else {
        await handleGuildMessage(message);
    }
});
