import { MessageFlags, type ButtonInteraction } from 'discord.js';
import { prisma } from '../../../../utils/core/database';
import { lsmsErrorEmbedGenerator, lsmsSuccessEmbedGenerator } from '../../../../utils/rp/lsms';

export async function handleLsmsDuty(interaction: ButtonInteraction): Promise<void> {
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

    if (!lsmsDutyData || !lsmsDutyData.dutyRoleId) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune donnée de service valide trouvée pour ce message.')],
        });
        return;
    }

    const memberRoles = interaction.member.roles;
    const hasDutyRole = memberRoles.cache.has(lsmsDutyData.dutyRoleId);
    let successMessage = '';

    if (hasDutyRole) {
        await memberRoles.remove(lsmsDutyData.dutyRoleId);
        successMessage = 'Vous avez quitté votre service.';
    } else {
        const hasOnCallRole = lsmsDutyData.onCallRoleId && memberRoles.cache.has(lsmsDutyData.onCallRoleId);

        const rolePromises = [];

        if (hasOnCallRole && lsmsDutyData.onCallRoleId) {
            rolePromises.push(memberRoles.remove(lsmsDutyData.onCallRoleId));
        }

        rolePromises.push(memberRoles.add(lsmsDutyData.dutyRoleId));
        await Promise.all(rolePromises);

        successMessage = 'Vous avez pris votre service.';
        if (hasOnCallRole) {
            successMessage += " Votre statut de semi service a été désactivé.";
        }
    }

    await interaction.editReply({ embeds: [lsmsSuccessEmbedGenerator(successMessage)] });
}
