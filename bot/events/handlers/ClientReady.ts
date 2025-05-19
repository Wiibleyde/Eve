import { ActivityType, Events } from 'discord.js';
import type { Event } from '../event';
import { deployCommands } from '../../commands/deploy-commands';
import { logger } from '../../..';
import { birthdayCron } from '../../../cron/birthdayCron';
import { statusCron } from '../../../cron/statusCron';

export const event: Event<Events.ClientReady> = {
    name: Events.ClientReady,
    once: true,
    execute: async (client) => {
        client.user?.setPresence({
            status: 'dnd',
            activities: [
                {
                    name: `le démarrage...`,
                    type: ActivityType.Watching,
                },
            ],
        });

        await deployCommands();

        birthdayCron.start();
        statusCron.start();

        logger.info(`Logged in as ${client.user?.tag}`);
    },
};
