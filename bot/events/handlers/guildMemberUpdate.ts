import { Events } from 'discord.js';
import type { Event } from '../event';
import { prisma } from '../../../utils/core/database';
import { client } from '../../bot';
import {
    lsmsDutyEmbedGenerator,
    lsmsDutyUpdateEmbedGenerator,
    lsmsOnCallUpdateEmbedGenerator,
} from '../../../utils/rp/lsms';
import { logger } from '../../..';

export const guildMemberUpdateEvent: Event<Events.GuildMemberUpdate> = {
    name: Events.GuildMemberUpdate,
    once: false,
    execute: async (oldMember, newMember) => {
        const startTime = Date.now();

        if (
            oldMember.roles.cache.size === newMember.roles.cache.size &&
            oldMember.roles.cache.every((role) => newMember.roles.cache.has(role.id))
        ) {
            return;
        }

        const roleChanges = newMember.roles.cache.filter((role) => !oldMember.roles.cache.has(role.id));
        const roleRemovals = oldMember.roles.cache.filter((role) => !newMember.roles.cache.has(role.id));

        if (roleChanges.size === 0 && roleRemovals.size === 0) {
            return;
        }

        const lsmsDutyData = await prisma.lsmsDutyManager.findMany({
            where: {
                AND: [
                    { guildId: newMember.guild.id },
                    {
                        OR: [
                            {
                                dutyRoleId: {
                                    in: [...roleChanges.map((role) => role.id), ...roleRemovals.map((role) => role.id)],
                                },
                            },
                            {
                                onCallRoleId: {
                                    in: [...roleChanges.map((role) => role.id), ...roleRemovals.map((role) => role.id)],
                                },
                            },
                        ],
                    },
                ],
            },
        });

        const dutyDataByMessage = new Map<string, typeof lsmsDutyData>();

        for (const dutyData of lsmsDutyData) {
            if (!dutyData.messageId) continue;
            if (!dutyDataByMessage.has(dutyData.messageId)) {
                dutyDataByMessage.set(dutyData.messageId, []);
            }
            dutyDataByMessage.get(dutyData.messageId)!.push(dutyData);
        }

        let actionPerformed = false;

        for (const [messageId, dutyDatas] of dutyDataByMessage.entries()) {
            const dutyData = dutyDatas[0];
            if (!dutyData) continue;
            const guild = await client.guilds.fetch(dutyData.guildId);
            const channel = guild.channels.cache.get(dutyData.channelId);
            if (!channel || !channel.isTextBased || !channel.isTextBased()) continue;

            let dutyMessage;
            try {
                dutyMessage = await channel.messages.fetch(messageId);
            } catch {
                continue;
            }
            if (!dutyMessage) continue;

            const embed = dutyMessage.embeds[0];
            if (!embed) continue;

            // Utilise le cache local des membres si possible
            let allMembers = guild.members.cache;
            if (allMembers.size < guild.memberCount) {
                try {
                    allMembers = await guild.members.fetch();
                } catch {
                    // fallback au cache si fetch échoue
                }
            }

            // Agréger tous les rôles duty et onCall à surveiller pour ce message
            const dutyRoleIds = dutyDatas.map((d) => d.dutyRoleId).filter(Boolean) as string[];
            const onCallRoleIds = dutyDatas.map((d) => d.onCallRoleId).filter(Boolean) as string[];

            const onDutyPeople = allMembers.filter(
                (member) => dutyRoleIds.some((roleId) => member.roles.cache.has(roleId)) && !member.user.bot
            );
            const onCallPeople = allMembers.filter(
                (member) => onCallRoleIds.some((roleId) => member.roles.cache.has(roleId)) && !member.user.bot
            );

            const { embed: newEmbed, actionRow } = lsmsDutyEmbedGenerator(
                Array.from(onDutyPeople.values()).map((member) => member.user),
                Array.from(onCallPeople.values()).map((member) => member.user)
            );
            await dutyMessage.edit({
                embeds: [newEmbed],
                components: [actionRow],
            });

            // Log pour chaque dutyData concerné
            for (const d of dutyDatas) {
                const logChannel = d.logsChannelId
                    ? guild.channels.cache.get(d.logsChannelId)
                    : null;
                if (logChannel && logChannel.isTextBased()) {
                    const isDutyRoleAdded =
                        d.dutyRoleId &&
                        roleChanges.some((role) => role.id === d.dutyRoleId);
                    const isOnCallRoleAdded =
                        d.onCallRoleId &&
                        roleChanges.some((role) => role.id === d.onCallRoleId);

                    const embed =
                        d.dutyRoleId &&
                        (isDutyRoleAdded ||
                            roleRemovals.some((role) => role.id === d.dutyRoleId))
                            ? lsmsDutyUpdateEmbedGenerator(newMember.user, !!isDutyRoleAdded)
                            : lsmsOnCallUpdateEmbedGenerator(newMember.user, !!isOnCallRoleAdded);
                    await logChannel.send({ embeds: [embed] });
                }
            }
            actionPerformed = true;
        }

        if (actionPerformed) {
            logger.info(
                `[${Date.now() - startTime}ms] Updated duty message due to role change for ${newMember.user.tag} (${newMember.id}) in ${newMember.guild.name} (${newMember.guild.id})`
            );
        }
        return;
    },
};
