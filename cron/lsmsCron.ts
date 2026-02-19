import { CronJob } from 'cron';
import { prisma } from '@utils/core/database';
import { logger } from '..';
import { client } from '@bot/bot';

// The cron job runs every day at 5:55AM
export const lsmsCron = new CronJob('55 5 * * *', async () => {
    const dutyManagers = await prisma.lsmsDutyManager.findMany();
    logger.info(`Found ${dutyManagers.length} LSMS duty managers.`);
    for (const dutyManager of dutyManagers) {
        // Remove every user from duty and onCall roles
        const guild = await client.guilds.fetch(dutyManager.guildId);
        const dutyRole = guild.roles.cache.find((role) => role.id === dutyManager.dutyRoleId);
        const onCallRole = guild.roles.cache.find((role) => role.id === dutyManager.onCallRoleId);
        const offRadioRole = guild.roles.cache.find((role) => role.id === dutyManager.offRadioRoleId);
        if (dutyRole) {
            dutyRole.members.forEach(async (member) => await member.roles.remove(dutyRole));
        }
        if (onCallRole) {
            onCallRole.members.forEach(async (member) => await member.roles.remove(onCallRole));
        }
        if (offRadioRole) {
            offRadioRole.members.forEach(async (member) => await member.roles.remove(offRadioRole));
        }
    }
});
