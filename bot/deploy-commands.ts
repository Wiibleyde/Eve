import { REST, Routes, type RESTPostAPIApplicationCommandsJSONBody } from 'discord.js';
import { commands } from './commands/command';
import { logger } from '..';
import { config } from '../utils/core/config';
import { messageContextMenuCommands } from './contextMenu/message/contextMenuMessage';
import { userContextMenuCommands } from './contextMenu/user/contextMenuUser';

export async function deployCommands(): Promise<void> {
    try {
        logger.info('Deploying commands...');

        const fullCommandList = [...commands, ...messageContextMenuCommands, ...userContextMenuCommands];

        // Create REST instance
        const rest = new REST({ version: '10' }).setToken(config.DISCORD_TOKEN);

        // Separate guild-specific commands from global commands
        const guildSpecificCommands = fullCommandList.filter(
            (command) => command.guildIds && command.guildIds.length > 0
        );
        const globalCommands = fullCommandList.filter((command) => !command.guildIds || command.guildIds.length === 0);

        // Deploy guild-specific commands
        if (guildSpecificCommands.length > 0) {
            const guildCommandsMap = new Map<string, RESTPostAPIApplicationCommandsJSONBody[]>();

            guildSpecificCommands.forEach((command) => {
                command.guildIds?.forEach((guildId) => {
                    if (!guildCommandsMap.has(guildId)) {
                        guildCommandsMap.set(guildId, []);
                    }
                    guildCommandsMap.get(guildId)!.push(command.data.toJSON());
                });
            });

            for (const [guildId, commandsData] of guildCommandsMap) {
                try {
                    await rest.put(Routes.applicationGuildCommands(config.DISCORD_CLIENT_ID, guildId), {
                        body: commandsData,
                    });
                    logger.info(
                        `Successfully deployed ${commandsData.length} guild-specific commands to guild: ${guildId}`
                    );
                } catch (error) {
                    logger.warn(`Failed to deploy commands to guild ${guildId}:`, error);
                }
            }
        }

        // Deploy global commands
        if (globalCommands.length > 0) {
            const globalCommandsData = globalCommands.map((command) => command.data.toJSON());
            await rest.put(Routes.applicationCommands(config.DISCORD_CLIENT_ID), {
                body: globalCommandsData,
            });
            logger.info(`Successfully deployed ${globalCommands.length} commands globally.`);
        }

        logger.info(
            `Total commands deployed: ${fullCommandList.length} (${globalCommands.length} global, ${guildSpecificCommands.length} guild-specific)`
        );
    } catch (error) {
        logger.error('Error deploying commands:', error);
        throw new Error(`Failed to deploy commands: ${error}`);
    }
}
