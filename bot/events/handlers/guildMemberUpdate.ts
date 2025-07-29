import { Events, GuildMember, type PartialGuildMember } from 'discord.js';
import type { Event } from '../event';
import { prisma } from '../../../utils/core/database';
import { client } from '../../bot';
import {
    lsmsDutyEmbedGenerator,
    lsmsDutyUpdateEmbedGenerator,
    lsmsOnCallUpdateEmbedGenerator,
    onCallUser,
    onDutyUser,
} from '../../../utils/rp/lsms';
import { logger } from '../../..';

// Buffer pour regrouper les modifications de rôles
const updateBuffer = new Map<
    string,
    {
        timeout: NodeJS.Timeout;
        oldMember: GuildMember | PartialGuildMember;
        newMember: GuildMember;
        allRoleChanges: Set<string>;
        allRoleRemovals: Set<string>;
    }
>();

const BUFFER_DELAY = 1500; // 1.5 secondes

const processRoleUpdate = async (
    oldMember: GuildMember | PartialGuildMember,
    newMember: GuildMember,
    roleChanges: Set<string>,
    roleRemovals: Set<string>
) => {
    const startTime = Date.now();

    const lsmsDutyData = await prisma.lsmsDutyManager.findMany({
        where: {
            AND: [
                { guildId: newMember.guild.id },
                {
                    OR: [
                        {
                            dutyRoleId: {
                                in: [...roleChanges, ...roleRemovals],
                            },
                        },
                        {
                            onCallRoleId: {
                                in: [...roleChanges, ...roleRemovals],
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

        // Ici, on sait qu'un rôle duty ou onCall a été modifié pour ce message
        // Tracker les utilisateurs selon les rôles modifiés
        for (const d of dutyDatas) {
            if (d.dutyRoleId && (roleChanges.has(d.dutyRoleId) || roleRemovals.has(d.dutyRoleId))) {
                onDutyUser(dutyData.guildId, newMember.user.id);
            }
            if (d.onCallRoleId && (roleChanges.has(d.onCallRoleId) || roleRemovals.has(d.onCallRoleId))) {
                onCallUser(dutyData.guildId, newMember.user.id);
            }
        }

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
            const logChannel = d.logsChannelId ? guild.channels.cache.get(d.logsChannelId) : null;
            if (logChannel && logChannel.isTextBased()) {
                const isDutyRoleAdded = d.dutyRoleId && roleChanges.has(d.dutyRoleId);
                const isDutyRoleRemoved = d.dutyRoleId && roleRemovals.has(d.dutyRoleId);
                const isOnCallRoleAdded = d.onCallRoleId && roleChanges.has(d.onCallRoleId);
                const isOnCallRoleRemoved = d.onCallRoleId && roleRemovals.has(d.onCallRoleId);

                // Gérer les changements de rôle de service
                if (isDutyRoleAdded || isDutyRoleRemoved) {
                    const dutyEmbed = lsmsDutyUpdateEmbedGenerator(newMember.user, !!isDutyRoleAdded);
                    await logChannel.send({ embeds: [dutyEmbed] });
                }

                // Gérer les changements de rôle de semi service
                if (isOnCallRoleAdded || isOnCallRoleRemoved) {
                    const onCallEmbed = lsmsOnCallUpdateEmbedGenerator(newMember.user, !!isOnCallRoleAdded);
                    await logChannel.send({ embeds: [onCallEmbed] });
                }
            }
        }
        actionPerformed = true;
    }

    if (actionPerformed) {
        logger.info(
            `[${Date.now() - startTime}ms] Updated duty message due to role change for ${newMember.user.tag} (${newMember.id}) in ${newMember.guild.name} (${newMember.guild.id})`
        );
    }
};

export const guildMemberUpdateEvent: Event<Events.GuildMemberUpdate> = {
    name: Events.GuildMemberUpdate,
    once: false,
    execute: async (oldMember, newMember) => {
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

        const userId = newMember.user.id;
        const guildId = newMember.guild.id;
        const bufferKey = `${guildId}-${userId}`;

        // Si un buffer existe déjà pour cet utilisateur, on l'annule et on accumule les changements
        if (updateBuffer.has(bufferKey)) {
            const existingBuffer = updateBuffer.get(bufferKey)!;
            clearTimeout(existingBuffer.timeout);

            // Accumuler les changements
            roleChanges.forEach((role) => existingBuffer.allRoleChanges.add(role.id));
            roleRemovals.forEach((role) => existingBuffer.allRoleRemovals.add(role.id));

            // Mettre à jour le newMember le plus récent
            existingBuffer.newMember = newMember;
        } else {
            // Créer un nouveau buffer
            updateBuffer.set(bufferKey, {
                timeout: setTimeout(() => {}, 0), // Sera remplacé ci-dessous
                oldMember,
                newMember,
                allRoleChanges: new Set(roleChanges.map((role) => role.id)),
                allRoleRemovals: new Set(roleRemovals.map((role) => role.id)),
            });
        }

        const buffer = updateBuffer.get(bufferKey)!;

        // Créer un nouveau timeout
        buffer.timeout = setTimeout(async () => {
            try {
                await processRoleUpdate(
                    buffer.oldMember,
                    buffer.newMember,
                    buffer.allRoleChanges,
                    buffer.allRoleRemovals
                );
            } catch (error) {
                logger.error('Error processing buffered role update:', error);
            } finally {
                updateBuffer.delete(bufferKey);
            }
        }, BUFFER_DELAY);

        return;
    },
};
