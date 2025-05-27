import { Events } from 'discord.js';
import type { Event } from '../event';
import { prisma } from '../../../utils/core/database';
import { client } from '../../bot';
import { lsmsDutyEmbedGenerator, lsmsDutyUpdateEmbedGenerator, lsmsOnCallUpdateEmbedGenerator } from '../../../utils/rp/lsms';
import { logger } from '../../..';

export const guildMemberUpdateEvent: Event<Events.GuildMemberUpdate> = {
    name: Events.GuildMemberUpdate,
    once: false,
    execute: async (oldMember, newMember) => {
        const startTime = Date.now();

        // Check if the member's roles have changed
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
                            { dutyRoleId: { in: [...roleChanges.map((role) => role.id), ...roleRemovals.map((role) => role.id)] } },
                            { onCallRoleId: { in: [...roleChanges.map((role) => role.id), ...roleRemovals.map((role) => role.id)] } },
                        ],
                    },
                ],
            },
        });

        let actionPerformed = false;

        for (const dutyData of lsmsDutyData) {
            const guild = await client.guilds.fetch(dutyData.guildId);
            // Update the message if either duty or oncall role was added OR removed
            const shouldUpdateMessage =
                (dutyData.dutyRoleId && (
                    roleChanges.some(role => role.id === dutyData.dutyRoleId) ||
                    roleRemovals.some(role => role.id === dutyData.dutyRoleId)
                )) ||
                (dutyData.onCallRoleId && (
                    roleChanges.some(role => role.id === dutyData.onCallRoleId) ||
                    roleRemovals.some(role => role.id === dutyData.onCallRoleId)
                ));

            if (shouldUpdateMessage) {
                const channel = guild.channels.cache.get(dutyData.channelId);
                if (channel && channel.isTextBased && channel.isTextBased()) {
                    if (dutyData.messageId) {
                        const dutyMessage = await channel.messages.fetch(dutyData.messageId);
                        if (dutyMessage) {
                            const embed = dutyMessage.embeds[0];
                            if (embed) {
                                const onDutyPeople = dutyData.dutyRoleId
                                    ? newMember.guild.members.cache.filter(member =>
                                        member.roles.cache.has(dutyData.dutyRoleId!) && !member.user.bot
                                    )
                                    : newMember.guild.members.cache.filter(() => false);

                                const onCallPeople = dutyData.onCallRoleId
                                    ? newMember.guild.members.cache.filter(member =>
                                        member.roles.cache.has(dutyData.onCallRoleId!) && !member.user.bot
                                    )
                                    : newMember.guild.members.cache.filter(() => false);
                                const { embed: newEmbed, actionRow } = lsmsDutyEmbedGenerator(onDutyPeople.map(member => member.user), onCallPeople.map(member => member.user));
                                await dutyMessage.edit({
                                    embeds: [newEmbed],
                                    components: [actionRow],
                                });
                                // Log to channel (use lsmsDutyUpdateEmbedGenerator or lsmsOnCallUpdateEmbedGenerator)
                                const logChannel = dutyData.logsChannelId ? guild.channels.cache.get(dutyData.logsChannelId) : null;
                                if (logChannel && logChannel.isTextBased()) {
                                    // Determine if role was added (true) or removed (false)
                                    const isDutyRoleAdded = dutyData.dutyRoleId && roleChanges.some(role => role.id === dutyData.dutyRoleId);
                                    const isOnCallRoleAdded = dutyData.onCallRoleId && roleChanges.some(role => role.id === dutyData.onCallRoleId);

                                    const embed = dutyData.dutyRoleId && (isDutyRoleAdded || roleRemovals.some(role => role.id === dutyData.dutyRoleId))
                                        ? lsmsDutyUpdateEmbedGenerator(newMember.user, !!isDutyRoleAdded)
                                        : lsmsOnCallUpdateEmbedGenerator(newMember.user, !!isOnCallRoleAdded);
                                    await logChannel.send({ embeds: [embed] });
                                }
                                actionPerformed = true;
                            }
                        }
                    }
                }
            }
        }
        if (actionPerformed) {
            logger.info(
                `[${Date.now() - startTime}ms] Updated duty message due to role change for ${newMember.user.tag} (${newMember.id}) in ${newMember.guild.name} (${newMember.guild.id})`
            );
        }
        return;
    },
};
