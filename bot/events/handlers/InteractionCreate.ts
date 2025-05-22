import { Events, MessageFlags } from 'discord.js';
import type { Event } from '../event';
import { logger } from '../../..';
import { commandsMap } from '../../commands/command';
import { buttons } from '../../buttons/buttons';
import { modals } from '../../modals/modals';

export const event: Event<Events.InteractionCreate> = {
    name: Events.InteractionCreate,
    once: false,
    execute: async (interaction) => {
        const startTime = Date.now();
        if (interaction.isChatInputCommand()) {
            const command = commandsMap.get(interaction.commandName);

            if (!command) {
                await interaction.reply({
                    content: "Cette commande n'existe pas.",
                    flags: [MessageFlags.Ephemeral],
                });
                return;
            }

            try {
                await command.execute(interaction);
            } catch (error) {
                logger.error(`Error executing command ${interaction.commandName}: ${error}`);
                await interaction.reply({
                    content: "Une erreur est survenue lors de l'exécution de la commande.",
                    flags: [MessageFlags.Ephemeral],
                });
            }
            logger.info(
                `[${Date.now() - startTime}ms] Command executed: ${interaction.commandName} by ${interaction.user.tag}`
            );
        } else if (interaction.isButton()) {
            const buttonHandler = buttons[interaction.customId];
            if (!buttonHandler) {
                await interaction.reply({
                    content: "Ce bouton n'existe pas.",
                    flags: [MessageFlags.Ephemeral],
                });
                return;
            }

            try {
                await buttonHandler(interaction);
            } catch (error) {
                logger.error(`Error executing button ${interaction.customId}: ${error}`);
                await interaction.reply({
                    content: "Une erreur est survenue lors de l'exécution du bouton.",
                    flags: [MessageFlags.Ephemeral],
                });
            }
            logger.info(
                `[${Date.now() - startTime}ms] Button executed: ${interaction.customId} by ${interaction.user.tag}`
            );
        } else if (interaction.isModalSubmit()) {
            const modalHandler = modals[interaction.customId];
            if (!modalHandler) {
                await interaction.reply({
                    content: "Ce modal n'existe pas.",
                    flags: [MessageFlags.Ephemeral],
                });
                return;
            }
            try {
                await modalHandler(interaction);
            } catch (error) {
                logger.error(`Error executing modal ${interaction.customId}: ${error}`);
                await interaction.reply({
                    content: "Une erreur est survenue lors de l'exécution du modal.",
                    flags: [MessageFlags.Ephemeral],
                });
            }
            logger.info(
                `[${Date.now() - startTime}ms] Modal executed: ${interaction.customId} by ${interaction.user.tag}`
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
