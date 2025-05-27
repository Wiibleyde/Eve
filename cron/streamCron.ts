import { CronJob } from 'cron';
import { checkStreamsUpdate } from '../utils/stream/twitch';
import { isMaintenanceMode } from '../utils/core/maintenance';

// This cron job is used to check the stream status of users every 10 seconds.
export const streamCron = new CronJob('*/10 * * * * *', async () => {
    if (isMaintenanceMode()) {
        // logger.warn('Maintenance mode is enabled, skipping stream check.');
        return;
    }

    await checkStreamsUpdate();
});
