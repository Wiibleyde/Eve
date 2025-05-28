import { MessageFlags, type ButtonInteraction } from 'discord.js';
import { prisma } from '../../../../utils/core/database';
import { lsmsErrorEmbedGenerator, lsmsSuccessEmbedGenerator } from '../../../../utils/rp/lsms';

export async function handleLsmsOnCall(interaction: ButtonInteraction): Promise<void> {
    if (!interaction.guildId || !interaction.inCachedGuild()) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator('Cette commande ne peut être utilisée que dans un serveur Discord.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });
    const messageId = interaction.message.id;

    const lsmsDutyData = await prisma.lsmsDutyManager.findFirst({
        where: {
            messageId,
            guildId: interaction.guildId,
        },
        select: {
            dutyRoleId: true,
            onCallRoleId: true,
        },
    });

    if (!lsmsDutyData || !lsmsDutyData.onCallRoleId) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune donnée de service valide trouvée pour ce message.')],
        });
        return;
    }

    const memberRoles = interaction.member.roles;
    const hasOnCallRole = memberRoles.cache.has(lsmsDutyData.onCallRoleId);
    let successMessage = '';

    if (hasOnCallRole) {
        // Retirer le rôle d'astreinte s'il est déjà présent
        await memberRoles.remove(lsmsDutyData.onCallRoleId);
        successMessage = 'Vous avez quitté votre astreinte.';
    } else {
        // Ajouter le rôle d'astreinte
        const hasDutyRole = lsmsDutyData.dutyRoleId && memberRoles.cache.has(lsmsDutyData.dutyRoleId);

        const rolePromises = [];

        if (hasDutyRole && lsmsDutyData.dutyRoleId) {
            rolePromises.push(memberRoles.remove(lsmsDutyData.dutyRoleId));
        }

        rolePromises.push(memberRoles.add(lsmsDutyData.onCallRoleId));
        await Promise.all(rolePromises);

        successMessage = 'Vous êtes maintenant en astreinte.';
        if (hasDutyRole) {
            successMessage += ' Votre service a été désactivé.';
        }
    }

    await interaction.editReply({ embeds: [lsmsSuccessEmbedGenerator(successMessage)] });
}
