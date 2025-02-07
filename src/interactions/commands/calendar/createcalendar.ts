import { logger } from '@/index';
import { findNextEvent, MCalendarComponent, upsertCalendarMessage } from '@/utils/calendar';
import { prisma } from '@/utils/database';
import { errorEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';
import {
    CommandInteraction,
    MessageFlags,
    PermissionFlagsBits,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
} from 'discord.js';
import { CalendarResponse, sync } from 'node-ical';

const icalMap = new Map<string, CalendarResponse>(); // Map<guildId, CalendarResponse>

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('createcalendar')
    .setDescription('Créer un calendrier (1 par serveur)')
    .addStringOption((option) => option.setName('url').setDescription('URL du calendrier').setRequired(true));

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });

    if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageEvents], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de changer la configuration."))],
        });
        return;
    }

    const url = interaction.options.get('url')?.value as string;

    if (
        await prisma.guildData.count({
            where: { AND: [{ guildId: interaction.guildId as string }, { calendarUrl: url }] },
        })
    ) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error('Ce calendrier est déjà enregistré.'))],
        });
        return;
    }

    await prisma.guildData.upsert({
        where: {
            guildId: interaction.guildId as string,
        },
        update: {
            calendarUrl: url,
        },
        create: {
            guildId: interaction.guildId as string,
            calendarUrl: url,
        },
    });

    await updateGuildCalendar(interaction.guildId as string);

    const calendar = icalMap.get(interaction.guildId as string);
    if (!calendar) {
        await interaction.editReply({ embeds: [errorEmbed(interaction, new Error('Calendrier introuvable.'))] });
        return;
    }
    const nextEvent = findNextEvent(calendar);
    if (!nextEvent) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error('Aucun événement à venir trouvé.'))],
        });
        return;
    }
    await upsertCalendarMessage(interaction.guildId as string, interaction.channelId, null, nextEvent);

    await interaction.editReply({ content: `Calendrier créé avec succès !` });
}

async function downloadCalendar(url: string): Promise<CalendarResponse> {
    const ics = await fetch(url).then((res) => res.text());
    return sync.parseICS(ics);
}

async function updateGuildCalendar(guildId: string) {
    const guildData = await prisma.guildData.findFirst({
        where: {
            AND: [
                {
                    guildId,
                },
                {
                    calendarUrl: {
                        not: null,
                    },
                },
            ],
        },
        select: {
            calendarUrl: true,
        },
    });

    if (!guildData) return;
    if (guildData.calendarUrl) {
        const calendar = await downloadCalendar(guildData.calendarUrl);
        if (!calendar) {
            logger.error(`Failed to download calendar for guild ${guildId}`);
            return;
        }
        icalMap.set(guildId, calendar);
    } else {
        icalMap.delete(guildId);
        logger.warn(`Previous calendar for guild ${guildId} was deleted.`);
    }
}

export async function initCalendars() {
    const guilds = await prisma.guildData.findMany({
        where: {
            calendarUrl: {
                not: null,
            },
        },
        select: {
            guildId: true,
        },
    });
    for (const guild of guilds) {
        await updateGuildCalendar(guild.guildId);
    }
}

export async function updateCalendars() {
    const guilds = await prisma.guildData.findMany({
        where: {
            calendarUrl: {
                not: null,
            },
        },
        select: {
            guildId: true,
            CalendarMessageData: {
                select: {
                    channelId: true,
                    messageId: true,
                },
            },
        },
    });
    for (const guild of guilds) {
        const nextEvent = findNextEvent(icalMap.get(guild.guildId) as CalendarResponse) as MCalendarComponent | null;
        // const event = areInEvent(icalMap.get(guild.guildId) as CalendarResponse) as MCalendarComponent | null
        upsertCalendarMessage(
            guild.guildId,
            guild.CalendarMessageData?.channelId as string,
            guild.CalendarMessageData?.messageId ?? null,
            nextEvent
        );
        // if(event) {
        //     await createDiscordEvent(guild.guildId, event)
        // }
    }
}
