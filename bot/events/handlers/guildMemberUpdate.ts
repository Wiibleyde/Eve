import { Events } from "discord.js";
import type { Event } from "../event";
import { prisma } from "../../../utils/core/database";
import { client } from "../../bot";

export const guildMemberUpdateEvent: Event<Events.GuildMemberUpdate> = {
    name: Events.GuildMemberUpdate,
    once: false,
    execute: async (oldMember, newMember) => {
        const startTime = Date.now();
        
        // Check if the member's roles have changed
        if (oldMember.roles.cache.size === newMember.roles.cache.size &&
            oldMember.roles.cache.every(role => newMember.roles.cache.has(role.id))) {
            return; 
        }

        const lsmsDutyData = await prisma.lsmsDutyManager.findMany({
            where: {
                AND: [
                    { guildId: newMember.guild.id },
                    { OR: [{ dutyRoleId: newMember.roles.cache.first()?.id }, { dutyRoleId: oldMember.roles.cache.first()?.id }] },
                ],
            },
        });
        for (const dutyData of lsmsDutyData) {
            // Check if the duty role or oncall role has changed
            if (dutyData.dutyRoleId !== newMember.roles.cache.first()?.id && dutyData.onCallRoleId !== newMember.roles.cache.first()?.id) {
                continue; 
            }
            const guildWhereUpdate = client.guilds.cache.get(dutyData.guildId);
            if (!guildWhereUpdate) continue;
            // const dutyRole = guildWhereUpdate.roles.cache.get(dutyData.dutyRoleId);
            // const onCallRole = guildWhereUpdate.roles.cache.get(dutyData.onCallRoleId);
        }
    },
};
