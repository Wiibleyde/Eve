import {
    ChatInputCommandInteraction,
    InteractionContextType,
    MessageFlags,
    PermissionFlagsBits,
    SlashCommandBuilder,
    type GuildBasedChannel,
} from 'discord.js';
import type { ICommand } from '../command';
import { prisma } from '../../../utils/core/database';
import { basicEmbedGenerator } from '../../utils/embeds';
import { hasPermission } from '../../../utils/permission';

export const config: ICommand = {
    data: new SlashCommandBuilder()
        .setName('config')
        .setDescription('Configure le bot')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('set')
                .setDescription('Configure une option')
                .addStringOption((option) =>
                    option
                        .setName('option')
                        .setDescription("L'option à configurer")
                        .setRequired(true)
                        .addChoices([
                            {
                                name: 'Salon des anniversaires',
                                value: 'birthdayChannel',
                            },
                            {
                                name: 'Salon des citations',
                                value: 'quoteChannel',
                            },
                        ])
                )
                .addChannelOption((option) =>
                    option.setName('channel').setDescription('Le salon à configurer').setRequired(true)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('get')
                .setDescription("Récupère la configuration d'une option")
                .addStringOption((option) =>
                    option
                        .setName('option')
                        .setDescription("L'option à récupérer")
                        .setRequired(true)
                        .addChoices([
                            {
                                name: 'Salon des anniversaires',
                                value: 'birthdayChannel',
                            },
                            {
                                name: 'Salon des citations',
                                value: 'quoteChannel',
                            },
                        ])
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('reset')
                .setDescription("Réinitialise la configuration d'une option")
                .addStringOption((option) =>
                    option
                        .setName('option')
                        .setDescription("L'option à réinitialiser")
                        .setRequired(true)
                        .addChoices([
                            {
                                name: 'Salon des anniversaires',
                                value: 'birthdayChannel',
                            },
                            {
                                name: 'Salon des citations',
                                value: 'quoteChannel',
                            },
                        ])
                )
        )
        .addSubcommand((subcommand) =>
            subcommand.setName('list').setDescription('Liste toutes les options de configuration')
        )
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels]))) {
            await interaction.editReply({
                embeds: [configEmbedGenerator().setDescription("Vous n'avez pas la permission de gérer les salons.")],
            });
            return;
        }

        const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();
        switch (subcommand) {
            case 'set': {
                const option = (interaction as ChatInputCommandInteraction).options.get('option')?.value as string;
                const channel = (interaction as ChatInputCommandInteraction).options.get('channel')?.channel as GuildBasedChannel;
                const actualDatabase = await prisma.config.findFirst({
                    where: {
                        AND: [
                            {
                                guildId: interaction.guildId as string,
                            },
                            {
                                key: option,
                            },
                        ],
                    },
                });
                if (actualDatabase) {
                    await prisma.config.update({
                        where: {
                            uuid: actualDatabase.uuid,
                        },
                        data: {
                            value: channel.id,
                        },
                    });
                } else {
                    await prisma.config.create({
                        data: {
                            guildId: interaction.guildId as string,
                            key: option,
                            value: channel.id,
                        },
                    });
                }
                const configEmbed = configEmbedGenerator()
                    .setDescription(`La configuration de l'option \`${option}\` a été mise à jour avec succès !`)
                    .addFields([
                        {
                            name: 'Salon',
                            value: `<#${channel.id}>`,
                        },
                    ]);
                await interaction.editReply({ embeds: [configEmbed] });
                break;
            }
            case 'get': {
                const option = (interaction as ChatInputCommandInteraction).options.get('option')?.value as string;
                const actualDatabase = await prisma.config.findFirst({
                    where: {
                        AND: [
                            {
                                guildId: interaction.guildId as string,
                            },
                            {
                                key: option,
                            },
                        ],
                    },
                });
                if (actualDatabase) {
                    const configEmbed = configEmbedGenerator()
                        .setDescription(`La configuration de l'option \`${option}\` est :`)
                        .addFields([
                            {
                                name: 'Salon',
                                value: `<#${actualDatabase.value}>`,
                            },
                        ]);
                    await interaction.editReply({ embeds: [configEmbed] });
                } else {
                    const configEmbed = configEmbedGenerator().setDescription(
                        `Aucune configuration trouvée pour l'option \`${option}\``
                    );
                    await interaction.editReply({ embeds: [configEmbed] });
                }
                break;
            }
            case 'reset': {
                const option = (interaction as ChatInputCommandInteraction).options.get('option')?.value as string;
                await prisma.config.deleteMany({
                    where: {
                        AND: [
                            {
                                guildId: interaction.guildId as string,
                            },
                            {
                                key: option,
                            },
                        ],
                    },
                });
                const configEmbed = configEmbedGenerator().setDescription(
                    `La configuration de l'option \`${option}\` a été réinitialisée avec succès !`
                );
                await interaction.editReply({ embeds: [configEmbed] });
                break;
            }
            case 'list': {
                const configEntries = await prisma.config.findMany({
                    where: {
                        guildId: interaction.guildId as string,
                    },
                });
                const configEmbed = configEmbedGenerator()
                    .setDescription(`Voici la liste des configurations pour votre serveur :`)
                    .addFields(
                        configEntries.map((entry) => ({
                            name: entry.key,
                            value: `<#${entry.value}>`,
                        }))
                    );
                await interaction.editReply({ embeds: [configEmbed] });
                break;
            }
        }
    },
};

function configEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - Configuration',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1374398072406151310/free-settings-icon-778-thumb.png?ex=682de773&is=682c95f3&hm=9f159c52bf30f8eeef4706a2c55f96c79639a7760a3a8abb6b151984923c290d&',
    });
}
