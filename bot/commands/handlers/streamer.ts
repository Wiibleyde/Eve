import {
    ApplicationCommandOptionType,
    ChannelType,
    ChatInputCommandInteraction,
    InteractionContextType,
    MessageFlags,
    SlashCommandBuilder,
} from 'discord.js';
import type { ICommand } from '../command';
import { getUserIdByLogin } from '../../../utils/stream/twitch';
import { errorEmbedGenerator, successEmbedGenerator } from '../../utils/embeds';
import { prisma } from '../../../utils/database';

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

                if (!channel || channel.type !== ChannelType.GuildText) {
                    await interaction.editReply({
                        embeds: [errorStreamerEmbedGenerator('Le salon doit être un salon textuel')],
                    });
                    return;
                }

                await prisma.stream.create({
                    data: {
                        twitchUserId: twitchUserId,
                        guildId: interaction.guildId as string,
                        channelId: channel.id,
                        roleId: role ? role.id : null,
                    },
                });

                // Logic to add the streamer to the list
                await interaction.editReply({
                    embeds: [
                        successStreamerEmbedGenerator(
                            `Streamer ${streamerName} ajouté aux notifications (Le message peut mettre 10s à arriver)`
                        ),
                    ],
                });
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
                // Logic to remove the streamer from the list
                await prisma.stream.delete({
                    where: {
                        uuid: existingStreamer.uuid,
                    },
                });
                await interaction.editReply({
                    embeds: [successStreamerEmbedGenerator(`Streamer ${streamerName} retiré des notifications`)],
                });
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
