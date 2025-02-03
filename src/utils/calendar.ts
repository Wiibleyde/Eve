import { EmbedBuilder, TextChannel } from 'discord.js';
import { CalendarComponent, CalendarResponse } from 'node-ical';
import { client } from '..';
import { prisma } from './database';

const exemptedEvents = ['Férié', 'Vacances', 'Jour férié'];

export type MCalendarComponent = CalendarComponent & {
    summary?: {
        val: string;
    };
    start: string | Date;
    end: string | Date;
};

export function findNextEvent(calendar: CalendarResponse): MCalendarComponent | null {
    let nextEvent: CalendarComponent | null = null;
    let nextEventTime: number = Infinity;
    const now = new Date().getTime();
    Object.keys(calendar).forEach((key) => {
        const event = calendar[key] as MCalendarComponent;
        const eventStart = new Date(event.start).getTime();
        if (event.summary?.val && !exemptedEvents.includes(event.summary.val)) {
            if (eventStart > now && eventStart < nextEventTime) {
                nextEvent = event;
                nextEventTime = eventStart;
            }
        }
    });
    return nextEvent;
}

// function areInEvent(calendar: CalendarResponse): MCalendarComponent | null {
//     let nextEvent: CalendarComponent | null = null
//     let nextEventTime: number = Infinity
//     const now = new Date().getTime()
//     Object.keys(calendar).forEach(key => {
//         const event = calendar[key] as MCalendarComponent
//         const eventStart = new Date(event.start).getTime()
//         const eventEnd = new Date(event.end).getTime()
//         if (event.summary?.val && !exemptedEvents.includes(event.summary.val)) {
//             if (eventStart < now && eventEnd > now && eventStart < nextEventTime) {
//                 nextEvent = event
//                 nextEventTime = eventStart
//             }
//         }
//     })
//     return nextEvent
// }

function prepareEmbed(nextEvent: MCalendarComponent | null): EmbedBuilder {
    if (!nextEvent) {
        return new EmbedBuilder()
            .setTitle('Prochain événement')
            .setColor(0xff0000)
            .setDescription(`Aucun événement à venir.`)
            .setTimestamp()
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() });
    }
    return new EmbedBuilder()
        .setTitle('Prochain événement')
        .setColor(0x00ff00)
        .setDescription(`**${nextEvent.summary?.val}**`)
        .addFields(
            {
                name: 'Début',
                value: `<t:${Math.floor(new Date(nextEvent.start).getTime() / 1000)}:R>`,
                inline: true,
            },
            {
                name: 'Fin',
                value: `<t:${Math.floor(new Date(nextEvent.end).getTime() / 1000)}:R>`,
                inline: true,
            }
        )
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() });
}

export async function upsertCalendarMessage(
    guildId: string,
    channelId: string,
    messageId: string | null,
    nextEvent: MCalendarComponent | null
): Promise<void> {
    if (messageId) {
        const message = await client.channels
            .fetch(channelId)
            .then((channel) => (channel as TextChannel).messages.fetch(messageId));
        await message.edit({ embeds: [prepareEmbed(nextEvent)] });
    } else {
        const message = await client.channels
            .fetch(channelId)
            .then((channel) => (channel as TextChannel).send({ embeds: [prepareEmbed(nextEvent)] }));
        await prisma.guildData.update({
            where: {
                guildId,
            },
            data: {
                CalendarMessageData: {
                    create: {
                        guildId: guildId,
                        channelId: channelId,
                        messageId: message.id,
                    },
                },
            },
        });
    }
}

// async function createDiscordEvent(guildId: string, eventToAdd: MCalendarComponent): Promise<void> {
//     const guild = await client.guilds.fetch(guildId);
//     const eventManager = guild.scheduledEvents;
//     if(!eventManager) {
//         logger.error(`Scheduled events are not enabled for guild ${guildId}`);
//         return;
//     }
//     const events = await eventManager.fetch();
//     for (const event of events) {
//         if (event[1].name === eventToAdd.summary?.val && event[1].scheduledStartAt?.getTime() === new Date(eventToAdd.start).getTime()) {
//             return;
//         }
//     }

//     const eventName = eventToAdd.summary?.val?.substring(0, 100) ?? "Événement";
//     const eventDescription = "Événement ajouté automatiquement par Eve.".substring(0, 1000);

//     eventManager.create({
//         name: eventName,
//         description: eventDescription,
//         scheduledStartTime: eventToAdd.start,
//         scheduledEndTime: eventToAdd.end,
//         privacyLevel: GuildScheduledEventPrivacyLevel.GuildOnly,
//         entityType: GuildScheduledEventEntityType.External,
//         entityMetadata: {
//             location: "Inconnu".substring(0, 100),
//         },
//         reason: "L'événement est dans le calendrier, ajouté automatiquement par Eve."
//     });
// }
