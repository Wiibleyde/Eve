import { ChatInputCommandInteraction, InteractionContextType, MessageFlags, PermissionFlagsBits, Role, SlashCommandBuilder, type GuildBasedChannel, type GuildTextBasedChannel } from "discord.js";
import type { ICommand } from "../../command";
import { hasPermission } from "../../../../utils/permission";
import { lsmsDutyEmbedGenerator, lsmsEmbedGenerator } from "../../../../utils/rp/lsms";
import { prisma } from "../../../../utils/core/database";

export const lsms: ICommand = {
    data: new SlashCommandBuilder()
        .setName("lsms")
        .setDescription("Commandes utiles pour le LSMS")
        .addSubcommand(subcommand =>
            subcommand
                .setName("duty")
                .setDescription("Créer un gestionnaire de service")
                .addRoleOption(option =>
                    option
                        .setName("duty")
                        .setDescription("Rôle à assigner pour le service")
                        .setRequired(true)
                )
                .addRoleOption(option =>
                    option
                        .setName("oncall")
                        .setDescription("Rôle à assigner pour l'appel")
                        .setRequired(true)
                )
                .addChannelOption(option =>
                    option
                        .setName("channel")
                        .setDescription("Salon où seront envoyés les logs des services")
                        .setRequired(true)
                )
        )
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels]))) {
            await interaction.editReply({
                embeds: [lsmsEmbedGenerator().setDescription("Vous n'avez pas la permission de gérer les salons.")],
            });
            return;
        }

        const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();
        switch (subcommand) {
            case "duty": {
                const dutyRole = interaction.options.get("duty", true).role as Role;
                const onCallRole = interaction.options.get("oncall", true).role as Role;
                const logsChannel = interaction.options.get("channel", true).channel as GuildBasedChannel;
                const interactionChannel = interaction.channel;
                if (!interactionChannel) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription("Impossible de déterminer le salon d'interaction.")],
                    });
                    return;
                }

                if (!logsChannel.isTextBased()) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription("Le salon doit être textuel.")],
                    });
                    return;
                }

                if (!interaction.guild) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription("Cette commande ne peut être utilisée que dans un serveur.")],
                    });
                    return;
                }

                if (!interaction.guild.roles.cache.has(dutyRole.id) || !interaction.guild.roles.cache.has(onCallRole.id)) {
                    await interaction.editReply({
                        embeds: [lsmsEmbedGenerator().setDescription("Les rôles doivent exister dans le serveur.")],
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
                        embeds: [lsmsEmbedGenerator().setDescription("Les rôles doivent être gérés par le bot et être inférieurs à son rôle le plus élevé.")],
                    });
                    return;
                }

                const usersWithDutyRole = interaction.guild.members.cache.filter(member => member.roles.cache.has(dutyRole.id));
                const usersWithOnCallRole = interaction.guild.members.cache.filter(member => member.roles.cache.has(onCallRole.id));
                const { embed, actionRow } = lsmsDutyEmbedGenerator(usersWithDutyRole.map(member => member.user), usersWithOnCallRole.map(member => member.user));
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

                break;
            }
            default:
                await interaction.editReply({
                    embeds: [lsmsEmbedGenerator().setDescription("Sous-commande inconnue.")],
                });
                break;
        }
    }
};
