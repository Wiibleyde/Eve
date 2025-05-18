import { REST, Routes } from 'discord.js';
import { commands } from './command';
import { logger } from '../..';
import { config } from '../../utils/config';

export async function deployCommands(): Promise<void> {
    try {
        logger.info('Deploying commands...');

        // Prepare the command data for registration
        const commandsData = commands.map(command => command.data.toJSON());

        // Create REST instance
        const rest = new REST({ version: '10' }).setToken(config.DISCORD_TOKEN);

        // Deploy globally - always deploy globally
        await rest.put(
            Routes.applicationCommands(config.DISCORD_CLIENT_ID),
            { body: commandsData },
        );
        logger.info('Successfully deployed commands globally.');
    } catch (error) {
        logger.error('Error deploying commands:', error);
        throw new Error(`Failed to deploy commands: ${error}`);
    }
}
