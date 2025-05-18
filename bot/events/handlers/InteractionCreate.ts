import { Events } from "discord.js";
import type { Event } from "../Event";
import { logger } from "../../..";
import { commandsMap } from "../../commands/command";

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
                    ephemeral: true,
                });
                return;
            }

            try {
                await command.execute(interaction);
            } catch (error) {
                logger.error(`Error executing command ${interaction.commandName}: ${error}`);
                await interaction.reply({
                    content: "Une erreur est survenue lors de l'exécution de la commande.",
                    ephemeral: true,
                });
            }
            logger.info(
                `[${Date.now() - startTime}ms] Command executed: ${interaction.commandName} by ${interaction.user.tag}`
            );
        }
    }
};