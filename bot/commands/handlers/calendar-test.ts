import {
    ChatInputCommandInteraction,
    InteractionContextType,
    MessageFlags,
    PermissionFlagsBits,
    SlashCommandBuilder,
} from 'discord.js';
import type { ICommand } from '../command';
import { prisma } from '../../../utils/core/database';
import { basicEmbedGenerator } from '../../utils/embeds';
import { hasPermission } from '../../../utils/permission';
import { Calendar } from '../../../utils/calendar';
import { logger } from '../../../index';

export const calendarTest: ICommand = {
    data: new SlashCommandBuilder()
        .setName('calendar-test')
        .setDescription('Teste le système de calendrier et affiche les événements détectés')
        .setContexts([InteractionContextType.Guild]),
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels]))) {
            await interaction.editReply({
                embeds: [basicEmbedGenerator().setDescription("Vous n'avez pas la permission de gérer les salons.")],
            });
            return;
        }

        // Get calendar URL from database
        const guildData = await prisma.guildData.findUnique({
            where: { guildId: interaction.guildId as string },
        });

        if (!guildData?.calendarUrl) {
            await interaction.editReply({
                embeds: [basicEmbedGenerator().setDescription("Aucun calendrier n'est configuré pour ce serveur.")],
            });
            return;
        }

        try {
            const cal = await Calendar.create(guildData.calendarUrl);
            const now = new Date();
            
            // Get all types of events
            const currentEvents = cal.getCurrentEvents();
            const upcomingEvents = cal.getUpcomingEvents();
            const soonEvents5 = cal.getEventsStartingSoon(5);
            const soonEvents15 = cal.getEventsStartingSoon(15);
            const soonEvents30 = cal.getEventsStartingSoon(30);

            const embed = basicEmbedGenerator()
                .setTitle('🧪 Test du Calendrier')
                .setDescription(`**Heure actuelle:** ${now.toLocaleString('fr-FR')}\n**URL:** ${guildData.calendarUrl.substring(0, 50)}...`);

            // Current events
            embed.addFields({
                name: `🔴 Événements en cours (${currentEvents.length})`,
                value: currentEvents.length > 0
                    ? currentEvents.map(e => {
                        const start = e.startDate?.toJSDate();
                        const end = e.endDate?.toJSDate();
                        return `• **${e.summary || 'Sans titre'}**\n  📅 ${start?.toLocaleString('fr-FR')} → ${end?.toLocaleString('fr-FR')}\n  🆔 ${e.uid || 'No UID'}`;
                    }).join('\n\n')
                    : 'Aucun événement en cours',
                inline: false,
            });

            // Upcoming events (next 3)
            embed.addFields({
                name: `⏭️ Prochains événements (${upcomingEvents.length})`,
                value: upcomingEvents.length > 0
                    ? upcomingEvents.slice(0, 3).map(e => {
                        const start = e.startDate?.toJSDate();
                        return `• **${e.summary || 'Sans titre'}**\n  📅 ${start?.toLocaleString('fr-FR')}\n  🆔 ${e.uid || 'No UID'}`;
                    }).join('\n\n')
                    : 'Aucun événement à venir',
                inline: false,
            });

            // Events starting soon (different windows)
            embed.addFields({
                name: `⏰ Événements dans 5 minutes (${soonEvents5.length})`,
                value: soonEvents5.length > 0
                    ? soonEvents5.map(e => {
                        const start = e.startDate?.toJSDate();
                        const minutesUntil = start ? Math.round((start.getTime() - now.getTime()) / 60000) : 0;
                        return `• **${e.summary || 'Sans titre'}**\n  ⏱️ Dans ${minutesUntil} minutes\n  🆔 ${e.uid || 'No UID'}`;
                    }).join('\n\n')
                    : 'Aucun événement',
                inline: false,
            });

            embed.addFields({
                name: `⏰ Événements dans 15 minutes (${soonEvents15.length})`,
                value: soonEvents15.length > 0
                    ? soonEvents15.map(e => {
                        const start = e.startDate?.toJSDate();
                        const minutesUntil = start ? Math.round((start.getTime() - now.getTime()) / 60000) : 0;
                        return `• **${e.summary || 'Sans titre'}**\n  ⏱️ Dans ${minutesUntil} minutes\n  🆔 ${e.uid || 'No UID'}`;
                    }).join('\n\n')
                    : 'Aucun événement',
                inline: false,
            });

            embed.addFields({
                name: `⏰ Événements dans 30 minutes (${soonEvents30.length})`,
                value: soonEvents30.length > 0
                    ? soonEvents30.map(e => {
                        const start = e.startDate?.toJSDate();
                        const minutesUntil = start ? Math.round((start.getTime() - now.getTime()) / 60000) : 0;
                        return `• **${e.summary || 'Sans titre'}**\n  ⏱️ Dans ${minutesUntil} minutes\n  🆔 ${e.uid || 'No UID'}`;
                    }).join('\n\n')
                    : 'Aucun événement',
                inline: false,
            });

            await interaction.editReply({ embeds: [embed] });

            // Log to console as well
            logger.info('Calendar test results:');
            logger.info(`Current events: ${currentEvents.length}`);
            logger.info(`Upcoming events: ${upcomingEvents.length}`);
            logger.info(`Events in 5 min: ${soonEvents5.length}`);
            logger.info(`Events in 15 min: ${soonEvents15.length}`);
            logger.info(`Events in 30 min: ${soonEvents30.length}`);

        } catch (error) {
            logger.error('Calendar test error:', error);
            await interaction.editReply({
                embeds: [basicEmbedGenerator().setDescription(`❌ Erreur: ${error}`)],
            });
        }
    },
};
