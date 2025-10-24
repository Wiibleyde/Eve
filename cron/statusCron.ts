import { CronJob } from 'cron';
import { ActivityType } from 'discord.js';
import { client } from '../bot/bot';
import { isMaintenanceMode } from '../utils/core/maintenance';

const possibleStatus: { name: string; type: ActivityType }[] = [
    { name: `les merveilles de ce monde.`, type: ActivityType.Watching },
    { name: `vos instructions.`, type: ActivityType.Listening },
    { name: `les données de mission.`, type: ActivityType.Watching },
    { name: `les étoiles.`, type: ActivityType.Watching },
];
const possibleHalloweenStatus: { name: string; type: ActivityType }[] = [
    { name: `la préparation des citrouilles. 🎃`, type: ActivityType.Competing },
    { name: `les fantômes... 👻`, type: ActivityType.Watching },
    { name: `Spooky Scary Skeletons`, type: ActivityType.Listening },
    { name: `des bonbons ou un sort ! 🍬`, type: ActivityType.Playing },
];
const possibleChristmasStatus: { name: string; type: ActivityType }[] = [
    { name: `l'emballage des cadeaux. 🎁`, type: ActivityType.Competing },
    { name: `les lutins. 🧝`, type: ActivityType.Watching },
    { name: `les chants de Noël`, type: ActivityType.Listening },
    { name: `le Père Noël. 🎅`, type: ActivityType.Playing },
];

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

let statusIndex: number = 0;

export const statusCron = new CronJob('0,10,20,30,40,50 * * * * *', () => {
    if (isMaintenanceMode()) {
        client.user?.setPresence({
            status: 'dnd',
            afk: true,
            activities: [
                {
                    name: 'la maintenance',
                    type: ActivityType.Watching,
                },
            ],
        });
        return;
    }

    let statusList: { name: string; type: ActivityType }[];
    if (areInPeriod(halloweenPeriod)) {
        statusList = possibleHalloweenStatus;
    } else if (areInPeriod(christmasPeriod)) {
        statusList = possibleChristmasStatus;
    } else {
        statusList = possibleStatus;
    }
    statusIndex = statusIndex % statusList.length;
    const status = statusList[statusIndex];
    statusIndex = (statusIndex + 1) % statusList.length;
    if (status) {
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
    }
});
