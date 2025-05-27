import { Guild, InteractionContextType, MessageFlags, Role, SlashCommandBuilder } from 'discord.js';
import type { ICommand } from '../command';
import { prisma } from '../../../utils/database';
import type { GuildData } from '@prisma/client';

export const debug: ICommand = {
    data: new SlashCommandBuilder()
        .setName('debug')
        .setDescription('Passer en mode debug sur ce serveur')
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });
        const server = interaction.guildId
            ? await interaction.client.guilds.fetch(interaction.guildId).catch(() => null)
            : null;

        if (!server) {
            await interaction.editReply({ content: 'Impossible de trouver le serveur' });
            return;
        }

        const guildId = interaction.guildId ?? '0';
        let serverConfig = await prisma.guildData.findFirst({
            where: { guildId },
        });

        if (!serverConfig) {
            serverConfig = await prisma.guildData.create({
                data: { guildId },
            });
        }

        const role = await getOrCreateDebugRole(server, serverConfig, interaction.client.user.id);

        if (!role) {
            await interaction.editReply({ content: 'Impossible de créer ou trouver le rôle de debug' });
            return;
        }

        if (serverConfig.debugRoleId !== role.id) {
            await prisma.guildData.update({
                where: { guildId },
                data: { debugRoleId: role.id },
            });
        }

        const member = server.members.cache.get(interaction.user.id);
        const hasRole = member?.roles.cache.has(role.id);

        if (hasRole) {
            await member?.roles.remove(role);
            await interaction.editReply({
                content: `Vous n'êtes plus en mode debug sur le serveur ${server.name}`,
            });
        } else {
            await member?.roles.add(role);
            await interaction.editReply({
                content: `Vous êtes maintenant en mode debug sur le serveur ${server.name}`,
            });
        }
    },
};

/**
 * Get an existing debug role or create a new one if it doesn't exist
 */
async function getOrCreateDebugRole(
    server: Guild,
    serverConfig: GuildData,
    botUserId: string
): Promise<Role | undefined> {
    if (serverConfig?.debugRoleId) {
        const existingRole = server?.roles.cache.get(serverConfig.debugRoleId) as Role;
        if (existingRole) return existingRole;
    }

    const botMember = server?.members.cache.get(botUserId);
    const botHighestRolePosition = botMember?.roles.highest.position || 0;

    return await server?.roles.create({
        name: 'Eve Debug',
        color: 'White',
        permissions: ['Administrator'],
        position: botHighestRolePosition - 1,
    });
}
