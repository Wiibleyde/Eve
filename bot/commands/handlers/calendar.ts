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
import { getStringOption, getSubcommand } from '../../utils/commandOptions';
import { Calendar } from '../../../utils/calendar';
import { logger } from '../../../index';

export const calendar: ICommand = {
    data: new SlashCommandBuilder()
        .setName('calendar')
        .setDescription('Gère le calendrier du serveur')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('set')
                .setDescription('Configure le calendrier avec un lien ICS')
                .addStringOption((option) =>
                    option
                        .setName('url')
                        .setDescription("L'URL du fichier ICS du calendrier")
                        .setRequired(true)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('remove')
                .setDescription('Supprime la configuration du calendrier')
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('refresh')
                .setDescription('Force la mise à jour du calendrier')
        )
        .setContexts([InteractionContextType.Guild]),
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels]))) {
            await interaction.editReply({
                embeds: [calendarEmbedGenerator().setDescription("Vous n'avez pas la permission de gérer les salons.")],
            });
            return;
        }

        const subcommand = getSubcommand(interaction);

        switch (subcommand) {
            case 'set': {
                const url = getStringOption(interaction, 'url', true);
                
                // Validate URL
                try {
                    new URL(url);
                } catch {
                    await interaction.editReply({
                        embeds: [calendarEmbedGenerator().setDescription("L'URL fournie n'est pas valide.")],
                    });
                    return;
                }

                // Test the ICS URL
                try {
                    await Calendar.create(url);
                } catch (error) {
                    logger.error('Failed to fetch calendar:', error);
                    await interaction.editReply({
                        embeds: [calendarEmbedGenerator().setDescription(
                            "Impossible de charger le calendrier depuis cette URL. Vérifiez qu'il s'agit bien d'un fichier ICS valide."
                        )],
                    });
                    return;
                }

                // Save to database
                const existingGuildData = await prisma.guildData.findUnique({
                    where: { guildId: interaction.guildId as string },
                });

                if (existingGuildData) {
                    await prisma.guildData.update({
                        where: { guildId: interaction.guildId as string },
                        data: { calendarUrl: url },
                    });
                } else {
                    await prisma.guildData.create({
                        data: {
                            guildId: interaction.guildId as string,
                            calendarUrl: url,
                        },
                    });
                }

                // Create the calendar message in the current channel
                const calendarEmbed = await generateCalendarEmbed(url);
                if (!interaction.channel || !('send' in interaction.channel)) {
                    await interaction.editReply({
                        embeds: [calendarEmbedGenerator().setDescription("Impossible d'envoyer un message dans ce canal.")],
                    });
                    return;
                }
                const message = await interaction.channel.send({ embeds: [calendarEmbed] });

                if (message) {
                    // Save message data
                    const botMessageData = await prisma.botMessageData.create({
                        data: {
                            guildId: interaction.guildId as string,
                            channelId: message.channelId,
                            messageId: message.id,
                        },
                    });

                    // Link to guild data
                    await prisma.guildData.update({
                        where: { guildId: interaction.guildId as string },
                        data: {
                            CalendarMessageDatauuid: botMessageData.uuid,
                            calendarMessageId: message.id,
                        },
                    });
                }

                const confirmEmbed = calendarEmbedGenerator()
                    .setDescription(`Le calendrier a été configuré avec succès !`)
                    .addFields([
                        {
                            name: 'URL du calendrier',
                            value: url,
                        },
                        {
                            name: 'Salon',
                            value: `<#${interaction.channelId}>`,
                        },
                    ]);
                await interaction.editReply({ embeds: [confirmEmbed] });
                break;
            }

            case 'remove': {
                const existingGuildData = await prisma.guildData.findUnique({
                    where: { guildId: interaction.guildId as string },
                    include: { CalendarMessageData: true },
                });

                if (!existingGuildData?.calendarUrl) {
                    await interaction.editReply({
                        embeds: [calendarEmbedGenerator().setDescription("Aucun calendrier n'est configuré pour ce serveur.")],
                    });
                    return;
                }

                // Delete the calendar message if it exists
                if (existingGuildData.CalendarMessageData) {
                    try {
                        const channel = await interaction.client.channels.fetch(existingGuildData.CalendarMessageData.channelId);
                        if (channel?.isTextBased()) {
                            const message = await channel.messages.fetch(existingGuildData.CalendarMessageData.messageId);
                            await message.delete();
                        }
                    } catch (error) {
                        logger.warn('Could not delete calendar message:', error);
                    }

                    // Delete the bot message data
                    await prisma.botMessageData.delete({
                        where: { uuid: existingGuildData.CalendarMessageDatauuid as string },
                    });
                }

                // Remove calendar from guild data
                await prisma.guildData.update({
                    where: { guildId: interaction.guildId as string },
                    data: {
                        calendarUrl: null,
                        calendarMessageId: null,
                        CalendarMessageDatauuid: null,
                    },
                });

                const confirmEmbed = calendarEmbedGenerator()
                    .setDescription('La configuration du calendrier a été supprimée avec succès !');
                await interaction.editReply({ embeds: [confirmEmbed] });
                break;
            }

            case 'refresh': {
                const existingGuildData = await prisma.guildData.findUnique({
                    where: { guildId: interaction.guildId as string },
                    include: { CalendarMessageData: true },
                });

                if (!existingGuildData?.calendarUrl) {
                    await interaction.editReply({
                        embeds: [calendarEmbedGenerator().setDescription("Aucun calendrier n'est configuré pour ce serveur.")],
                    });
                    return;
                }

                // Update the calendar message
                if (existingGuildData.CalendarMessageData) {
                    try {
                        const channel = await interaction.client.channels.fetch(existingGuildData.CalendarMessageData.channelId);
                        if (channel?.isTextBased()) {
                            const message = await channel.messages.fetch(existingGuildData.CalendarMessageData.messageId);
                            const calendarEmbed = await generateCalendarEmbed(existingGuildData.calendarUrl);
                            await message.edit({ embeds: [calendarEmbed] });
                        }
                    } catch (error) {
                        logger.error('Failed to refresh calendar message:', error);
                        await interaction.editReply({
                            embeds: [calendarEmbedGenerator().setDescription("Erreur lors de la mise à jour du message du calendrier.")],
                        });
                        return;
                    }
                }

                const confirmEmbed = calendarEmbedGenerator()
                    .setDescription('Le calendrier a été mis à jour avec succès !');
                await interaction.editReply({ embeds: [confirmEmbed] });
                break;
            }
        }
    },
};

function calendarEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - Calendrier',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1374398072406151310/free-settings-icon-778-thumb.png?ex=682de773&is=682c95f3&hm=9f159c52bf30f8eeef4706a2c55f96c79639a7760a3a8abb6b151984923c290d&',
    });
}

export async function generateCalendarEmbed(calendarUrl: string) {
    const embed = calendarEmbedGenerator().setTitle('📅 Calendrier des événements');

    try {
        const cal = await Calendar.create(calendarUrl);
        const currentEvents = cal.getCurrentEvents();
        const upcomingEvents = cal.getUpcomingEvents();

        if (currentEvents.length > 0) {
            const currentEvent = currentEvents[0];
            const startDate = currentEvent?.startDate?.toJSDate();
            const endDate = currentEvent?.endDate?.toJSDate();
            embed.addFields({
                name: '🔴 Événement en cours',
                value: `**${currentEvent?.summary || 'Sans titre'}**\n` +
                    `📍 ${currentEvent?.location || 'Lieu non spécifié'}\n` +
                    `⏰ Début : ${startDate ? `<t:${Math.floor(startDate.getTime() / 1000)}:F>` : 'N/A'}\n` +
                    `⏰ Fin : ${endDate ? `<t:${Math.floor(endDate.getTime() / 1000)}:F>` : 'N/A'}\n` +
                    (currentEvent?.description ? `📝 ${currentEvent.description.substring(0, 100)}${currentEvent.description.length > 100 ? '...' : ''}` : ''),
                inline: false,
            });
        } else {
            embed.addFields({
                name: '🔴 Événement en cours',
                value: 'Aucun événement en cours',
                inline: false,
            });
        }

        if (upcomingEvents.length > 0) {
            const nextEvent = upcomingEvents[0];
            const startDate = nextEvent?.startDate?.toJSDate();
            const endDate = nextEvent?.endDate?.toJSDate();
            embed.addFields({
                name: '⏭️ Prochain événement',
                value: `**${nextEvent?.summary || 'Sans titre'}**\n` +
                    `📍 ${nextEvent?.location || 'Lieu non spécifié'}\n` +
                    `⏰ Début : ${startDate ? `<t:${Math.floor(startDate.getTime() / 1000)}:R>` : 'N/A'}\n` +
                    `⏰ Fin : ${endDate ? `<t:${Math.floor(endDate.getTime() / 1000)}:t>` : 'N/A'}\n` +
                    (nextEvent?.description ? `📝 ${nextEvent.description.substring(0, 100)}${nextEvent.description.length > 100 ? '...' : ''}` : ''),
                inline: false,
            });
        } else {
            embed.addFields({
                name: '⏭️ Prochain événement',
                value: 'Aucun événement à venir',
                inline: false,
            });
        }

        embed.setFooter({ text: `Dernière mise à jour : ${new Date().toLocaleString('fr-FR')}` });
    } catch (error) {
        logger.error('Failed to generate calendar embed:', error);
        embed.setDescription('❌ Erreur lors du chargement du calendrier');
    }

    return embed;
}
