import { REST, Routes } from 'discord.js';
import { config } from '@/config';
import { commands, devCommands } from '@/interactions/commands';
import { logger } from '@/index';
import { client } from '.';
import { contextMessageMenus, contextUserMenus } from './interactions/contextMenus';

const commandsData = Object.values(commands).map((command) => command.data);
const devCommandsData = Object.values(devCommands).map((command) => command.data);
const contextCommandsData = [
    ...Object.values(contextMessageMenus).map((command) => command.data),
    ...Object.values(contextUserMenus).map((command) => command.data),
];

const rest = new REST().setToken(config.DISCORD_TOKEN);

/**
 * Deploys the commands to the Discord application.
 *
 * This function uploads the command data to the Discord API, making the commands available globally.
 * It logs the progress and any errors encountered during the process.
 *
 * @returns {Promise<void>} A promise that resolves when the commands have been successfully deployed.
 *
 * @throws Will log an error message if the deployment fails.
 */
export async function deployCommands(): Promise<void> {
    try {
        logger.info(`Chargement des commandes globales (${commandsData.length + contextCommandsData.length})...`);

        await rest.put(Routes.applicationCommands(config.DISCORD_CLIENT_ID), {
            body: [...commandsData, ...contextCommandsData],
        });

        logger.info(`${commandsData.length + contextCommandsData.length} commandes chargées avec succès.`);
    } catch (error) {
        logger.error(error as string);
    }
}

/**
 * Deploys development commands to a specified guild.
 *
 * @param guildId - The ID of the guild where the commands will be deployed.
 * @returns A promise that resolves when the deployment is complete.
 *
 * @throws Will log an error if the deployment fails or if the guild cannot be found.
 *
 * @example
 * ```typescript
 * await deployDevCommands('123456789012345678');
 * ```
 */
export async function deployDevCommands(guildId: string): Promise<void> {
    try {
        logger.info(
            `Chargement des commandes de développement pour la guilde ${guildId} (${devCommandsData.length})...`
        );

        await rest.put(Routes.applicationGuildCommands(config.DISCORD_CLIENT_ID, guildId), {
            body: devCommandsData,
        });

        const guild = client.guilds.cache.get(guildId);
        if (!guild) {
            logger.error(`Impossible de trouver la guilde ${guildId}`);
            return;
        }
        logger.info('Commandes chargées avec succès pour la guilde :', guild.name, guild.id);
    } catch (error) {
        logger.error(error as string);
    }
}
