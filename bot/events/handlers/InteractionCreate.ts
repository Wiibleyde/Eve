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

// Helper function to handle interactions with error handling and logging
async function handleInteraction(
    interaction: ChatInputCommandInteraction | ButtonInteraction | ModalSubmitInteraction,
    handlerFn: (() => Promise<void>) | undefined,
    interactionType: string,
    interactionId: string
) {
    const startTime = Date.now();

    if (!handlerFn) {
        await interaction.reply({
            content: `Ce ${interactionType} n'existe pas.`,
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    try {
        await handlerFn();
    } catch (error) {
        logger.error(`Error executing ${interactionType} ${interactionId}: ${error}`);
        await interaction.reply({
            content: `Une erreur est survenue lors de l'exécution du ${interactionType}.`,
            flags: [MessageFlags.Ephemeral],
        });
    }

    logger.info(
        `[${Date.now() - startTime}ms] ${interactionType} executed: ${interactionId} by ${interaction.user.tag}`
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
            const buttonHandler = buttons[interaction.customId];
            await handleInteraction(
                interaction,
                buttonHandler ? () => buttonHandler(interaction) : undefined,
                'bouton',
                interaction.customId
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
