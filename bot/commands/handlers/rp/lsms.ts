import {
    ChatInputCommandInteraction,
    InteractionContextType,
    MessageFlags,
    PermissionFlagsBits,
    Role,
    SlashCommandBuilder,
    type GuildBasedChannel,
    type GuildTextBasedChannel,
} from 'discord.js';
import type { ICommand } from '../../command';
import { hasPermission } from '../../../../utils/permission';
import { lsmsDutyEmbedGenerator, lsmsEmbedGenerator, updateRadioMessage } from '../../../../utils/rp/lsms';
import { prisma } from '../../../../utils/core/database';
import { config } from '@utils/core/config';
import { getSubcommand, getRoleOption, getChannelOption, getStringOption } from '../../../utils/commandOptions';

export const lsms: ICommand = {
    data: new SlashCommandBuilder()
        .setName('lsms')
        .setDescription('[LSMS] Commandes utiles pour le LSMS')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('addduty')
                .setDescription('Créer un gestionnaire de service')
                .addRoleOption((option) =>
                    option.setName('duty').setDescription('Rôle à assigner pour le service').setRequired(true)
                )
                .addRoleOption((option) =>
                    option.setName('oncall').setDescription("Rôle à assigner pour l'appel").setRequired(true)
                )
                .addChannelOption((option) =>
                    option
                        .setName('logchannel')
                        .setDescription('Salon où seront envoyés les logs des services')
                        .setRequired(true)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('removeduty')
                .setDescription('Supprimer un gestionnaire de service')
                .addStringOption((option) =>
                    option
                        .setName('messageid')
                        .setDescription('ID du message du gestionnaire de service à supprimer')
                        .setRequired(true)
                )
        )
        .addSubcommand((subcommand) => subcommand.setName('radio').setDescription('Créer un gestionnaire de radio'))
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    guildIds: ['872119977946263632', config.EVE_HOME_GUILD], // This command is available in all guilds
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels]))) {
            await interaction.editReply({
                embeds: [lsmsEmbedGenerator().setDescription("Vous n'avez pas la permission de gérer les salons.")],
            });
            return;
        }

        const subcommand = getSubcommand(interaction);
        switch (subcommand) {
            case 'addduty': {
                const dutyRole = getRoleOption(interaction, 'duty', true) as Role;
                const onCallRole = getRoleOption(interaction, 'oncall', true) as Role;
                const logsChannel = getChannelOption(interaction, 'logchannel', true) as GuildBasedChannel;
                const interactionChannel = interaction.channel;
                if (!interactionChannel) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription("Impossible de déterminer le salon d'interaction."),
                        ],
                    });
                    return;
                }

                if (!logsChannel.isTextBased()) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription('Le salon doit être textuel.')],
                    });
                    return;
                }

                if (!interaction.guild) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription(
                                'Cette commande ne peut être utilisée que dans un serveur.'
                            ),
                        ],
                    });
                    return;
                }

                if (
                    !interaction.guild.roles.cache.has(dutyRole.id) ||
                    !interaction.guild.roles.cache.has(onCallRole.id)
                ) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription('Les rôles doivent exister dans le serveur.')],
                    });
                    return;
                }

                const botHighestRole = interaction.guild.members.me?.roles.highest;
                if (
                    !botHighestRole ||
                    dutyRole.position > botHighestRole.position ||
                    onCallRole.position > botHighestRole.position
                ) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription(
                                'Les rôles doivent être gérés par le bot et être inférieurs à son rôle le plus élevé.'
                            ),
                        ],
                    });
                    return;
                }

                const usersWithDutyRole = interaction.guild.members.cache.filter((member) =>
                    member.roles.cache.has(dutyRole.id)
                );
                const usersWithOnCallRole = interaction.guild.members.cache.filter((member) =>
                    member.roles.cache.has(onCallRole.id)
                );
                const { embed, actionRow } = lsmsDutyEmbedGenerator(
                    usersWithDutyRole.map((member) => member.user),
                    usersWithOnCallRole.map((member) => member.user)
                );
                const message = await (interactionChannel as GuildTextBasedChannel).send({
                    embeds: [embed],
                    components: [actionRow],
                });

                await prisma.lsmsDutyManager.create({
                    data: {
                        guildId: interaction.guild.id,
                        channelId: interactionChannel.id,
                        logsChannelId: logsChannel.id,
                        dutyRoleId: dutyRole.id,
                        onCallRoleId: onCallRole.id,
                        messageId: message.id,
                    },
                });

                await interaction.editReply({
                    embeds: [lsmsEmbedGenerator().setDescription('Le gestionnaire de service a été créé.')],
                });

                break;
            }
            case 'removeduty': {
                const messageId = getStringOption(interaction, 'messageid', true);
                if (!interaction.guild) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription(
                                'Cette commande ne peut être utilisée que dans un serveur.'
                            ),
                        ],
                    });
                    return;
                }

                // Recherche du duty manager dans la base de données
                const dutyManager = await prisma.lsmsDutyManager.findFirst({
                    where: {
                        guildId: interaction.guild.id,
                        messageId: messageId,
                    },
                });

                if (!dutyManager) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription(
                                'Aucun gestionnaire de service trouvé avec cet ID de message.'
                            ),
                        ],
                    });
                    return;
                }

                // Suppression du message si possible
                try {
                    const channel = await interaction.guild.channels.fetch(dutyManager.channelId);
                    if (channel?.isTextBased()) {
                        const msg = await channel.messages.fetch(messageId);
                        if (msg) await msg.delete();
                    }
                } catch {
                    // Ignore si le message n'existe plus ou n'est pas accessible
                }

                // Suppression de la base de données
                await prisma.lsmsDutyManager.deleteMany({
                    where: {
                        guildId: interaction.guild.id,
                        messageId: messageId,
                    },
                });

                await interaction.editReply({
                    embeds: [lsmsEmbedGenerator().setDescription('Le gestionnaire de service a été supprimé.')],
                });

                break;
            }
            case 'radio': {
                const interactionChannel = interaction.channel;

                if (!interactionChannel || !interactionChannel.isTextBased()) {
                    await interaction.editReply({
                        embeds: [
                            lsmsEmbedGenerator().setDescription(
                                'Cette commande doit être utilisée dans un salon textuel.'
                            ),
                        ],
                    });
                    return;
                }

                const { embed, components } = updateRadioMessage([]);

                const textChannel = interactionChannel as GuildTextBasedChannel;

                await textChannel.send({
                    embeds: [embed],
                    components,
                });

                await interaction.editReply({
                    embeds: [lsmsEmbedGenerator().setDescription('Le gestionnaire de radios a été créé.')],
                });
                break;
            }
            default:
                await interaction.editReply({
                    embeds: [lsmsEmbedGenerator().setDescription('Sous-commande inconnue.')],
                });
                break;
        }
    },
};
