import { InteractionContextType, MessageFlags, Role, SlashCommandBuilder } from "discord.js";
import type { ICommand } from "../command";
import { prisma } from "../../../utils/database";

export const debug: ICommand = {
    data: new SlashCommandBuilder()
        .setName("debug")
        .setDescription("Passer en mode debug sur ce serveur")
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });
        const server = interaction.guildId
            ? await interaction.client.guilds.fetch(interaction.guildId).catch(() => null)
            : null;
        const serverConfig = await prisma.guildData.findFirst({
            where: {
                guildId: interaction.guildId ?? "0",
            },
        });
        let role: Role | undefined;
        if (!serverConfig) {
            await prisma.guildData.create({
                data: {
                    guildId: interaction.guildId ?? "0",
                },
            });

            // Get the bot's highest role position
            const botMember = server?.members.cache.get(interaction.client.user.id);
            const botHighestRolePosition = botMember?.roles.highest.position || 0;

            role = await server?.roles.create({
                name: 'Eve Debug',
                color: 'White',
                permissions: ['Administrator'],
                position: botHighestRolePosition - 1, // Set position higher than bot's highest role
            });

            await prisma.guildData.update({
                where: {
                    guildId: interaction.guildId ?? "0",
                },
                data: {
                    debugRoleId: role?.id,
                },
            });
        } else {
            role = server?.roles.cache.get(serverConfig?.debugRoleId as string) as Role;
            if (!role) {
                // Get the bot's highest role position
                const botMember = server?.members.cache.get(interaction.client.user.id);
                const botHighestRolePosition = botMember?.roles.highest.position || 0;

                role = (await server?.roles.create({
                    name: 'Eve Debug',
                    color: 'White',
                    permissions: ['Administrator'],
                    position: botHighestRolePosition - 1,
                })) as Role;
                await prisma.guildData.update({
                    where: {
                        guildId: interaction.guildId ?? "0",
                    },
                    data: {
                        debugRoleId: role?.id,
                    },
                });
                await server?.members.cache.get(interaction.user.id)?.roles.add(role);
                await interaction.editReply({
                    content: `Vous êtes maintenant en mode debug sur le serveur ${server?.name}`,
                });
                return;
            }
        }

        if (!server) {
            await interaction.editReply({ content: 'Impossible de trouver le serveur' });
            return;
        }
        const userRoles = await server.members.fetch(interaction.user.id).then(async (member) => await member.roles.cache);
        if (role && userRoles?.has(role.id)) {
            await server?.members.cache.get(interaction.user.id)?.roles.remove(role);
            await interaction.editReply({ content: `Vous n'êtes plus en mode debug sur le serveur ${server?.name}` });
            return;
        }
        if (!role) {
            await interaction.editReply({ content: 'Impossible de trouver le rôle de debug' });
            return;
        }
        await server?.members.cache.get(interaction.user.id)?.roles.add(role);
        await interaction.editReply({ content: `Vous êtes maintenant en mode debug sur le serveur ${server?.name}` });

    }
};
