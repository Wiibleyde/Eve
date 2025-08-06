import { ActivityType, Events } from 'discord.js';
import type { Event } from '../event';
import { deployCommands } from '../../deploy-commands';
import { logger } from '../../..';
import { initStreamsUpdate } from '../../../utils/stream/twitch';
import { initMpThreads } from '../../../utils/mpManager';
import { birthdayCron } from '@cron/birthdayCron';
import { statusCron } from '@cron/statusCron';
import { streamCron } from '@cron/streamCron';
import { lsmsCron } from '@cron/lsmsCron';

export const clientReadyEvent: Event<Events.ClientReady> = {
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

        await initMpThreads();

        await initStreamsUpdate();

        await deployCommands();

        birthdayCron.start();
        statusCron.start();
        streamCron.start();
        lsmsCron.start();

        logger.info(`Logged in as ${client.user?.tag}`);
    },
};
