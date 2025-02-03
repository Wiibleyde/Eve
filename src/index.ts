import { ActivityType, Client, EmbedBuilder, Events, GatewayIntentBits, Partials } from 'discord.js';
import { deployCommands, deployDevCommands } from '@/deploy-commands';
import { config } from '@/config';
import { Logger } from '@/utils/logger';
import { CronJob } from 'cron';
import { prisma } from '@/utils/database';
import { initAi } from '@/utils/intelligence';
import { maintenance } from '@/interactions/commands/dev/maintenance';
import { initCalendars, updateCalendars } from './interactions/commands/calendar/createcalendar';
import { Player } from 'discord-player';
import { YoutubeiExtractor } from 'discord-player-youtubei';
import { initMpThreads } from './utils/mpManager';
import readline from 'node:readline';

export const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: 'eve> ',
});

rl.on('close', () => {
    process.exit(0);
});

rl.prompt();

enum Commands {
    EXIT = 'exit',
    QUIT = 'quit',
    HELP = 'help',
}

const commands = Object.values(Commands);

export const logger = new Logger();
logger.initLevels();

export const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.GuildMembers,
        GatewayIntentBits.GuildVoiceStates,
        GatewayIntentBits.GuildMessageReactions,
        GatewayIntentBits.GuildMessageTyping,
        GatewayIntentBits.DirectMessages,
        GatewayIntentBits.DirectMessageReactions,
        GatewayIntentBits.DirectMessageTyping,
        GatewayIntentBits.MessageContent,
    ],
    partials: [Partials.User, Partials.Channel, Partials.Message, Partials.GuildMember],
});

export const player = new Player(client);
player.extractors.register(YoutubeiExtractor, { streamOptions: { highWaterMark: 1 << 25 } });

client.once(Events.ClientReady, async () => {
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
    const devGuild = config.EVE_HOME_GUILD;
    await deployDevCommands(devGuild);

    logger.info(`Connecté en tant que ${client.user?.tag}!`);

    // Load events
    import('./events/player');
    import('./events/discord');
});

/**
 * A CronJob that runs daily at midnight to check for users' birthdays.
 *
 * This job performs the following tasks:
 * 1. Retrieves the current date and extracts the day and month.
 * 2. Queries the database for users whose birthdays match the current day and month.
 * 3. Fetches all guilds the bot is a member of.
 * 4. For each user with a birthday, checks if the user is a member of each guild.
 * 5. If the user is a member of the guild, retrieves the guild's configuration to find the birthday channel.
 * 6. Sends a birthday message to the configured channel in the guild.
 *
 * The birthday message includes:
 * - A title "Joyeux anniversaire !"
 * - A description mentioning the user and their age (if birth date is available)
 * - A color set to white
 * - A timestamp
 * - A footer with the bot's name and avatar
 *
 * Logs errors if:
 * - The user is not found in the guild
 * - The birthday channel is not found in the guild configuration
 * - The birthday channel is not found in the guild
 */
const birthdayCron = new CronJob('0 0 0 * * *', async () => {
    const today = new Date();
    const todayDay = today.getDate();
    const todayMonth = today.getMonth() + 1;

    const todayBirthdays: {
        uuid: string;
        userId: string;
        birthDate: Date | null;
        quizGoodAnswers: number;
        quizBadAnswers: number;
    }[] =
        await prisma.$queryRaw`SELECT uuid, userId, birthDate, quizGoodAnswers, quizBadAnswers FROM GlobalUserData WHERE EXTRACT(DAY FROM birthDate) = ${todayDay} AND EXTRACT(MONTH FROM birthDate) = ${todayMonth}`;

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
                        const embed = new EmbedBuilder()
                            .setTitle('Joyeux anniversaire !')
                            .setDescription(
                                `Joyeux anniversaire <@${birthday.userId}> (${birthday.birthDate ? new Date().getFullYear() - new Date(birthday.birthDate).getFullYear() : ''} ans) ! 🎉🎂`
                            )
                            .setColor(0xffffff)
                            .setTimestamp()
                            .setFooter({
                                text: `Eve – Toujours prête à vous aider.`,
                                iconURL: client.user?.displayAvatarURL(),
                            });
                        await channel.send({ embeds: [embed] });
                    } else {
                        logger.error(`Channel ${guildConfig.value} not found in guild ${guild.id}`);
                    }
                } else {
                    logger.error(`Channel not found in guild ${guild.id}`);
                }
            } else {
                logger.error(`Member ${birthday.userId} not found in guild ${guild.id}`);
            }
        }
    }
});
birthdayCron.start();

const possibleStatus: { name: string; type: ActivityType }[] = [
    { name: `les merveilles de ce monde.`, type: ActivityType.Watching },
    { name: `vos instructions.`, type: ActivityType.Listening },
    { name: `les anomalies environnementales.`, type: ActivityType.Competing },
    { name: `les données de mission.`, type: ActivityType.Watching },
];
const possibleHalloweenStatus: { name: string; type: ActivityType }[] = [
    { name: `la préparation des citrouilles. 🎃`, type: ActivityType.Competing },
    { name: `les fantômes... 👻`, type: ActivityType.Watching },
    { name: `Spooky Scary Skeletons`, type: ActivityType.Listening },
    { name: `les bonbons ou un sort ! 🍬`, type: ActivityType.Playing },
];
const possibleChristmasStatus: { name: string; type: ActivityType }[] = [
    { name: `l'emballage des cadeaux. 🎁`, type: ActivityType.Competing },
    { name: `les lutins. 🧝`, type: ActivityType.Watching },
    { name: `les chants de Noël`, type: ActivityType.Listening },
    { name: `le Père Noël. 🎅`, type: ActivityType.Playing },
];
let statusIndex = 0;
const halloweenPeriod: { start: Date; end: Date } = {
    start: new Date(new Date().getFullYear(), 9, 24),
    end: new Date(new Date().getFullYear(), 10, 7),
};
const christmasPeriod: { start: Date; end: Date } = {
    start: new Date(new Date().getFullYear(), 11, 1),
    end: new Date(new Date().getFullYear(), 11, 25),
};

const areInPeriod = (period: { start: Date; end: Date }) => {
    const today = new Date();
    return today >= period.start && today <= period.end;
};

/**
 * A CronJob that updates the bot's presence status every 10 seconds.
 *
 * The status is determined based on the current period (e.g., Halloween, Christmas)
 * or a general status if no special period is active. If the bot is in maintenance mode,
 * it sets a specific maintenance status.
 *
 * The statuses are cycled through from predefined lists of possible statuses.
 *
 * @cron '0,10,20,30,40,50 * * * * *' - Runs every 10 seconds.
 *
 * @async
 * @function
 * @returns {Promise<void>} - A promise that resolves when the presence is updated.
 */
const statusCron = new CronJob('0,10,20,30,40,50 * * * * *', () => {
    if (maintenance) {
        client.user?.setPresence({
            status: 'idle',
            activities: [
                {
                    name: `la maintenance...`,
                    type: ActivityType.Competing,
                },
            ],
        });
        return;
    }
    let status: { name: string; type: ActivityType };
    if (areInPeriod(halloweenPeriod)) {
        status = possibleHalloweenStatus[statusIndex];
    } else if (areInPeriod(christmasPeriod)) {
        status = possibleChristmasStatus[statusIndex];
    } else {
        status = possibleStatus[statusIndex];
    }
    client.user?.setPresence({
        status: 'online',
        afk: false,
        activities: [
            {
                name: status.name,
                type: status.type,
            },
        ],
    });
    statusIndex++;
    if (statusIndex >= possibleStatus.length) {
        statusIndex = 0;
    }
});
statusCron.start();

const calendarCron = new CronJob('0 0 0 * * *', async () => {
    await initCalendars();
});
calendarCron.start();

// CronJob to update the calendar events every 10 minutes
const calendarEventsCron = new CronJob('0 */10 * * * *', async () => {
    await updateCalendars();
});
calendarEventsCron.start();

process.on('SIGINT', async () => {
    logger.info('Ctrl-C détécté, déconnexion...');
    await prisma.$disconnect();
    await client.destroy();
    logger.info('Déconnecté, arrêt du bot...');
    rl.close();
    process.exit(0);
});

initMpThreads();
initAi();
initCalendars();

rl.on('line', (line) => {
    const command = line.trim().toLowerCase();

    if (command === '') {
        rl.prompt();
        return;
    }

    switch (command) {
        case Commands.EXIT:
        case Commands.QUIT:
            process.kill(process.pid, 'SIGINT');
            break;
        case Commands.HELP:
            logger.info('Commandes disponibles :');
            logger.info("- exit, quit : Quitter l'application");
            logger.info('- help : Afficher la liste des commandes disponibles');
            break;
        default:
            logger.warn('Commande inconnue :', command);
            logger.info("Tapez 'help' pour afficher la liste des commandes disponibles");
            break;
    }
});

rl.on('tab', (line) => {
    const hits = commands.filter((c) => c.startsWith(line.trim().toLowerCase()));
    if (hits.length === 1) {
        rl.write(null, { ctrl: true, name: 'u' });
        rl.write(hits[0]);
        rl.prompt(true);
    } else if (hits.length > 1) {
        logger.info('Suggestions :', hits.join(', '));
        rl.prompt();
    } else {
        rl.prompt();
    }
});

(rl as readline.Interface & { input: NodeJS.ReadableStream }).input.on(
    'keypress',
    (char: string, key: { name: string }) => {
        if (key && key.name === 'tab') {
            const line = rl.line.trim().toLowerCase();
            const hits = commands.filter((c) => c.startsWith(line));
            if (hits.length === 0) {
                return false; // Prevent tab character if no suggestions are found
            }
            rl.emit('tab', rl.line);
        }
    }
);

client.login(config.DISCORD_TOKEN);
