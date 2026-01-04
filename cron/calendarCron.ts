import { CronJob } from "cron";
import {
	GuildScheduledEventEntityType,
	GuildScheduledEventPrivacyLevel,
} from "discord.js";
import { logger } from "..";
import { client } from "../bot/bot";
import { generateCalendarEmbed } from "../bot/commands/handlers/calendar";
import { Calendar, type CalendarEventLike } from "../utils/calendar";
import { prisma } from "../utils/core/database";
import { isMaintenanceMode } from "../utils/core/maintenance";

// Store created Discord events to avoid duplicates
const createdDiscordEvents = new Map<string, Set<string>>(); // guildId -> Set<eventUid>

export const calendarCron = new CronJob(
	"*/5 * * * *", // Runs every 5 minutes
	async () => {
		if (isMaintenanceMode()) {
			return;
		}

		try {
			// Get all guilds with calendar configured
			const guildsWithCalendar = await prisma.guildData.findMany({
				where: {
					calendarUrl: {
						not: null,
					},
				},
				include: {
					CalendarMessageData: true,
				},
			});

			for (const guildData of guildsWithCalendar) {
				if (!guildData.calendarUrl) continue;

				try {
					// Update the calendar message
					if (guildData.CalendarMessageData) {
						const guild = await client.guilds.fetch(guildData.guildId);
						const channel = await guild.channels.fetch(
							guildData.CalendarMessageData.channelId,
						);

						if (channel?.isTextBased() && "messages" in channel) {
							const message = await channel.messages.fetch(
								guildData.CalendarMessageData.messageId,
							);
							const calendarEmbed = await generateCalendarEmbed(
								guildData.calendarUrl,
							);
							await message.edit({ embeds: [calendarEmbed] });
						}
					}

					// Check for events starting in 30 minutes and create Discord scheduled events
					const cal = await Calendar.create(guildData.calendarUrl);
					const soonEvents = cal.getEventsStartingSoon(5);

					if (soonEvents.length > 0) {
						for (const evt of soonEvents) {
							logger.debug(
								`Soon event: ${evt.summary} at ${evt.startDate?.toJSDate()}`,
							);
						}
					}

					const guild = await client.guilds.fetch(guildData.guildId);

					// Initialize the set for this guild if not exists
					if (!createdDiscordEvents.has(guildData.guildId)) {
						createdDiscordEvents.set(guildData.guildId, new Set());
					}
					const guildCreatedEvents =
						createdDiscordEvents.get(guildData.guildId) ?? new Set<string>();

					for (const event of soonEvents) {
						// Skip if we already created this event
						if (!event.uid || guildCreatedEvents.has(event.uid)) {
							logger.debug(
								`Skipping event ${event.summary} (uid: ${event.uid}, already created: ${guildCreatedEvents.has(event.uid || "")})`,
							);
							continue;
						}

						try {
							await createDiscordScheduledEvent(guild.id, event);
							guildCreatedEvents.add(event.uid);
						} catch (error) {
							logger.error(
								`Failed to create Discord event for ${event.summary}:`,
								error,
							);
						}
					}

					// Clean up old event UIDs from the map (older than 24 hours)
					// This is done by checking the current events and keeping only recent ones
					const currentAndUpcoming = [
						...cal.getCurrentEvents(),
						...cal.getUpcomingEvents(),
					];
					const validUids = new Set(
						currentAndUpcoming.map((e) => e.uid).filter(Boolean),
					);

					// Keep only UIDs that still exist in the calendar
					for (const uid of Array.from(guildCreatedEvents)) {
						if (!validUids.has(uid)) {
							guildCreatedEvents.delete(uid);
						}
					}
				} catch (error) {
					logger.error(
						`Failed to update calendar for guild ${guildData.guildId}:`,
						error,
					);
				}
			}
		} catch (error) {
			logger.error("Error in calendar cron:", error);
		}
	},
);

async function createDiscordScheduledEvent(
	guildId: string,
	event: CalendarEventLike,
) {
	if (!event.startDate || !event.endDate) {
		logger.warn(
			"Event missing start or end date, skipping Discord event creation",
		);
		return;
	}

	const guild = await client.guilds.fetch(guildId);
	const startDate = event.startDate.toJSDate();
	const endDate = event.endDate.toJSDate();

	logger.debug(
		`Attempting to create Discord event: ${event.summary}, start: ${startDate}, end: ${endDate}`,
	);

	// Check if a similar event already exists
	const existingEvents = await guild.scheduledEvents.fetch();
	const duplicate = existingEvents.find(
		(e) =>
			e.name === event.summary &&
			e.scheduledStartAt?.getTime() === startDate.getTime(),
	);

	if (duplicate) {
		logger.debug(`Discord event already exists for ${event.summary}, skipping`);
		return;
	}

	logger.debug(`Creating Discord scheduled event for ${event.summary}`);
	if (event.summary && event.summary.length > 100) {
		event.summary = event.summary.substring(0, 100); // Discord limit is 100 chars
	}
	await guild.scheduledEvents.create({
		name: event.summary || "Événement",
		description: event.description?.substring(0, 1000) || undefined, // Discord limit is 1000 chars
		scheduledStartTime: startDate,
		scheduledEndTime: endDate,
		privacyLevel: GuildScheduledEventPrivacyLevel.GuildOnly,
		entityType: event.location
			? GuildScheduledEventEntityType.External
			: GuildScheduledEventEntityType.External, // Using External for all events
		entityMetadata: event.location
			? {
					location: event.location.substring(0, 100), // Discord limit is 100 chars
				}
			: {
					location: "À déterminer",
				},
	});
}
