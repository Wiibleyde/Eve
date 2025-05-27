import {
    ChannelType,
    ChatInputCommandInteraction,
    InteractionContextType,
    MessageFlags,
    SlashCommandBuilder,
} from 'discord.js';
import type { ICommand } from '../command';
import { getUserIdByLogin, initSingleStreamUpdate, removeStreamFromCache } from '../../../utils/stream/twitch';
import { errorEmbedGenerator, successEmbedGenerator } from '../../utils/embeds';
import { prisma } from '../../../utils/core/database';
import { logger } from '../../..';
import { hasPermission } from '../../../utils/permission';

export const streamer: ICommand = {
    data: new SlashCommandBuilder()
        .setName('streamer')
        .setDescription('Gérer les notifications de stream')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('add')
                .setDescription('Ajouter un streamer à la liste de notifications')
                .addStringOption((option) =>
                    option.setName('streamer').setDescription('Nom du streamer à ajouter').setRequired(true)
                )
                .addChannelOption((option) =>
                    option
                        .setName('channel')
                        .setDescription('Salon à mentionner lors de la notification')
                        .setRequired(true)
                )
                .addRoleOption((option) =>
                    option
                        .setName('role')
                        .setDescription('Rôle à mentionner lors de la notification')
                        .setRequired(false)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('remove')
                .setDescription('Retirer un streamer de la liste de notifications')
                .addStringOption((option) =>
                    option.setName('streamer').setDescription('Nom du streamer à retirer').setRequired(true)
                )
        )
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    execute: async (interaction) => {
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

        if (!(await hasPermission(interaction, []))) {
            await interaction.editReply({
                embeds: [errorStreamerEmbedGenerator("Vous n'avez pas la permission de gérer les streamers")],
            });
            return;
        }

        const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();
        switch (subcommand) {
            case 'add': {
                const streamerName = interaction.options.get('streamer', true).value as string;
                const channel = interaction.options.get('channel', true).channel;
                const role = interaction.options.get('role')?.role;

                // Prepare the insert data in the database
                const twitchUserId = await getUserIdByLogin(streamerName);
                if (!twitchUserId) {
                    await interaction.editReply({ embeds: [errorStreamerEmbedGenerator('Nom de streamer invalide')] });
                    return;
                }

                // Check if the streamer is already in the database
                const existingStreamer = await prisma.stream.findFirst({
                    where: {
                        AND: [{ twitchUserId: twitchUserId }, { guildId: interaction.guildId as string }],
                    },
                });
                if (existingStreamer) {
                    await interaction.editReply({
                        embeds: [errorStreamerEmbedGenerator('Le streamer est déjà dans la liste sur le serveur')],
                    });
                    return;
                }

                if (!channel || channel.type !== ChannelType.GuildText && channel.type !== ChannelType.GuildAnnouncement) {
                    await interaction.editReply({
                        embeds: [errorStreamerEmbedGenerator('Le salon doit être un salon textuel')],
                    });
                    return;
                }

                // Insert the streamer data in the database with role ID if provided
                await prisma.stream.create({
                    data: {
                        twitchUserId,
                        guildId: interaction.guildId as string,
                        channelId: channel.id,
                        roleId: role ? role.id : null,
                    },
                });

                // Initialisation immédiate du streamer ajouté avec envoi forcé de notification
                try {
                    await initSingleStreamUpdate(twitchUserId, true);
                    await interaction.editReply({
                        embeds: [
                            successStreamerEmbedGenerator(
                                `Streamer ${streamerName} ajouté aux notifications (Configuration terminée)`
                            ),
                        ],
                    });
                } catch (error) {
                    console.error("Erreur lors de l'initialisation du streamer:", error);
                    await interaction.editReply({
                        embeds: [
                            errorStreamerEmbedGenerator(
                                `Streamer ${streamerName} ajouté, mais une erreur est survenue lors de l'initialisation`
                            ),
                        ],
                    });
                }
                break;
            }
            case 'remove': {
                const streamerName = interaction.options.get('streamer', true).value as string;
                const twitchUserId = await getUserIdByLogin(streamerName);
                if (!twitchUserId) {
                    await interaction.editReply({ embeds: [errorStreamerEmbedGenerator('Nom de streamer invalide')] });
                    return;
                }
                // Check if the streamer is in the database
                const existingStreamer = await prisma.stream.findFirst({
                    where: {
                        AND: [{ twitchUserId: twitchUserId }, { guildId: interaction.guildId as string }],
                    },
                });
                if (!existingStreamer) {
                    await interaction.editReply({
                        embeds: [errorStreamerEmbedGenerator("Le streamer n'est pas dans la liste sur le serveur")],
                    });
                    return;
                }

                try {
                    // Suppression immédiate du message avant la suppression de la base de données
                    if (existingStreamer.messageId) {
                        const { deleteStreamMessage } = await import('../../../bot/utils/stream');
                        await deleteStreamMessage(existingStreamer);
                    }

                    // Logic to remove the streamer from the list
                    await prisma.stream.delete({
                        where: {
                            uuid: existingStreamer.uuid,
                        },
                    });

                    // Suppression immédiate du streamer du cache
                    await removeStreamFromCache(twitchUserId);

                    await interaction.editReply({
                        embeds: [successStreamerEmbedGenerator(`Streamer ${streamerName} retiré des notifications`)],
                    });
                } catch (error) {
                    logger.error(`Error removing streamer ${streamerName}:`, error);
                    await interaction.editReply({
                        embeds: [
                            errorStreamerEmbedGenerator(
                                `Une erreur est survenue lors de la suppression du streamer ${streamerName}`
                            ),
                        ],
                    });
                }
                break;
            }
        }
    },
};

function errorStreamerEmbedGenerator(reason: string) {
    return errorEmbedGenerator(reason).setAuthor({
        name: 'Eve - Streamers',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1374731886328287264/logo-twitch_578229-259.png?ex=682f1e56&is=682dccd6&hm=d4f74f6d9760dd699c2da474afe28fa195217270692fb1de38b5bcc214b1c367&',
    });
}

function successStreamerEmbedGenerator(reason: string) {
    return successEmbedGenerator(reason).setAuthor({
        name: 'Eve - Streamers',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1374731886328287264/logo-twitch_578229-259.png?ex=682f1e56&is=682dccd6&hm=d4f74f6d9760dd699c2da474afe28fa195217270692fb1de38b5bcc214b1c367&',
    });
}
