import {
    ButtonInteraction,
    ChatInputCommandInteraction,
    Events,
    MessageFlags,
    ModalSubmitInteraction,
} from 'discord.js';
import type { Event } from '../event';
import { logger } from '../../..';
import { commandsMap } from '../../commands/command';
import { buttons } from '../../buttons/buttons';
import { modals } from '../../modals/modals';
import { isMaintenanceMode } from '../../../utils/maintenance';
import { warningEmbedGenerator } from '../../utils/embeds';
import { config } from '../../../utils/config';

// Function to parse button ID with arguments
function parseCustomId(customId: string): { baseId: string; args: string | null } {
    const parts = customId.split('--');
    return {
        baseId: parts[0] ?? '',
        args: parts.length > 1 ? (parts[1] ?? null) : null,
    };
}

// Helper function to handle interactions with error handling and logging
async function handleInteraction(
    interaction: ChatInputCommandInteraction | ButtonInteraction | ModalSubmitInteraction,
    handlerFn: (() => Promise<void>) | undefined,
    interactionType: string,
    interactionId: string
) {
    const startTime = Date.now();

    if (isMaintenanceMode() && interaction.user.id !== config.OWNER_ID) {
        logger.warn(`Maintenance mode is enabled, interaction ${interactionType} ${interactionId} ignored.`);
        if (interaction.isRepliable() && !interaction.replied && !interaction.deferred) {
            try {
                await interaction.reply({
                    embeds: [
                        warningEmbedGenerator(
                            `Le bot est actuellement en mode maintenance. Veuillez réessayer plus tard.`
                        ),
                    ],
                    flags: [MessageFlags.Ephemeral],
                });
            } catch (replyError) {
                logger.error(`Failed to reply to ${interactionType} in maintenance mode: ${replyError}`);
            }
        }
        return;
    }

    // Get context information
    const guildName = interaction.guild ? interaction.guild.name : 'DM';
    const guildId = interaction.guild ? interaction.guild.id : 'DM';
    const channelName = interaction.channel
        ? 'name' in interaction.channel
            ? interaction.channel.name
            : 'unknown'
        : 'unknown';
    const channelId = interaction.channelId || 'unknown';

    // Get arguments for commands
    let argumentsInfo = '';
    if (interaction.isChatInputCommand()) {
        const options = interaction.options.data;
        if (options.length > 0) {
            argumentsInfo = ` with args: ${options
                .map((opt) => `${opt.name}=${opt.value ?? '[complex value]'}`)
                .join(', ')}`;
        }
    }

    if (!handlerFn) {
        // Only reply if the interaction hasn't been replied to yet
        if (interaction.isRepliable() && !interaction.replied && !interaction.deferred) {
            try {
                await interaction.reply({
                    content: `Ce ${interactionType} n'existe pas.`,
                    flags: [MessageFlags.Ephemeral],
                });
            } catch (replyError) {
                logger.error(`Failed to reply to unknown ${interactionType}: ${replyError}`);
            }
        }
        logger.warn(
            `Unknown ${interactionType} attempted: ${interactionId} by ${interaction.user.tag} (<@${interaction.user.id}>) in ${guildName} (${guildId}), channel: ${channelName} (<#${channelId}>)`
        );
        return;
    }

    try {
        await handlerFn();
    } catch (error) {
        logger.error(`Error executing ${interactionType} ${interactionId}${argumentsInfo}: ${error}`);

        // Only attempt to reply if the interaction is repliable and hasn't been replied to yet
        if (interaction.isRepliable() && !interaction.replied && !interaction.deferred) {
            try {
                await interaction.reply({
                    content: `Une erreur est survenue lors de l'exécution du ${interactionType}.`,
                    flags: [MessageFlags.Ephemeral],
                });
            } catch (replyError: unknown) {
                // If the error is not because the interaction was already replied to, log it
                if (
                    typeof replyError === 'object' &&
                    replyError !== null &&
                    'code' in replyError &&
                    (replyError as { code?: string }).code !== 'InteractionAlreadyReplied'
                ) {
                    logger.error(`Failed to send error response: ${replyError}`);
                }
            }
        }
        return;
    }

    logger.info(
        `[${Date.now() - startTime}ms] ${interactionType}: ${interactionId}${argumentsInfo} by ${interaction.user.tag} (<@${interaction.user.id}>) in ${guildName} (${guildId}), channel: ${channelName} (<#${channelId}>)`
    );
}

export const interactionCreateEvent: Event<Events.InteractionCreate> = {
    name: Events.InteractionCreate,
    once: false,
    execute: async (interaction) => {
        if (interaction.isChatInputCommand()) {
            const command = commandsMap.get(interaction.commandName);
            await handleInteraction(
                interaction,
                command ? () => command.execute(interaction) : undefined,
                'commande',
                interaction.commandName
            );
        } else if (interaction.isButton()) {
            // Parse the button ID to extract base ID and arguments
            const { baseId, args } = parseCustomId(interaction.customId);
            const buttonHandler = buttons[baseId];

            await handleInteraction(
                interaction,
                buttonHandler ? () => buttonHandler(interaction) : undefined,
                'bouton',
                baseId + (args ? ` (args: ${args})` : '')
            );
        } else if (interaction.isModalSubmit()) {
            const modalHandler = modals[interaction.customId];
            await handleInteraction(
                interaction,
                modalHandler ? () => modalHandler(interaction) : undefined,
                'modal',
                interaction.customId
            );
        } else {
            logger.warn(`Interaction not handled: ${interaction.type}`);
            if (interaction.isRepliable()) {
                await interaction.reply({
                    content: "Cette interaction n'est pas gérée.",
                    flags: [MessageFlags.Ephemeral],
                });
            }
            return;
        }
    },
};
