import { CronJob } from 'cron';
import { logger } from '..';
import { prisma } from '../utils/core/database';
import { client } from '../bot/bot';
import { birthdayEmbedGenerator } from '../bot/commands/handlers/birthday';
import { isMaintenanceMode } from '../utils/core/maintenance';

export const birthdayCron = new CronJob(
    '0 0 * * *', // Runs every day at midnight
    async () => {
        if (isMaintenanceMode()) {
            logger.warn('Maintenance mode is enabled, skipping birthday check.');
            return;
        }

        logger.info('Checking birthdays...');
        const today = new Date();
        const todayDay = today.getDate();
        const todayMonth = today.getMonth() + 1; // Months are zero-based

        const todayBirthdays: {
            uuid: string;
            userId: string;
            birthDate: Date | null;
        }[] =
            await prisma.$queryRaw`SELECT uuid, userId, birthDate FROM GlobalUserData WHERE EXTRACT(DAY FROM birthDate) = ${todayDay} AND EXTRACT(MONTH FROM birthDate) = ${todayMonth}`;

        const botGuilds = await client.guilds.fetch().then((guilds) => guilds.map((guild) => guild));
        for (const birthday of todayBirthdays) {
            for (const guild of botGuilds) {
                const usersOnGuild = await client.guilds.fetch(guild.id).then((guild) => guild.members.fetch());
                const member = usersOnGuild.get(birthday.userId);
                if (member) {
                    const guildConfig = await prisma.config.findFirst({
                        where: {
                            guildId: guild.id,
                            key: 'birthdayChannel',
                        },
                    });
                    if (guildConfig) {
                        const fullGuild = await client.guilds.fetch(guild.id);
                        const channel = fullGuild.channels.cache.get(guildConfig.value);
                        if (channel && channel.isTextBased()) {
                            const birthdayEmbed = birthdayEmbedGenerator();
                            birthdayEmbed.setDescription(
                                `Joyeux anniversaire <@${member.id}>  (${birthday.birthDate ? new Date().getFullYear() - new Date(birthday.birthDate).getFullYear() : ''} ans) ! 🎉🎂`
                            );
                            birthdayEmbed.setColor('#FF69B4');

                            await channel.send({
                                content: `<@${member.id}>`,
                                embeds: [birthdayEmbed],
                            });
                        } else {
                            logger.error(`Le channel d'anniversaire n'est pas valide dans le serveur ${guild.name}`);
                        }
                    } else {
                        logger.warn(`Aucune configuration d'anniversaire trouvée pour le serveur ${guild.name}`);
                    }
                }
            }
        }
        logger.info('Birthday check completed.');
    }
);
