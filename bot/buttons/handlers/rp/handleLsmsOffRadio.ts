import { MessageFlags, type ButtonInteraction } from 'discord.js';
import { prisma } from '../../../../utils/core/database';
import { lsmsErrorEmbedGenerator, lsmsSuccessEmbedGenerator } from '../../../../utils/rp/lsms';

export async function handleLsmsOffRadio(interaction: ButtonInteraction): Promise<void> {
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
            offRadioRoleId: true,
        },
    });

    if (!lsmsDutyData || !lsmsDutyData.offRadioRoleId) {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune donnée de off radio valide trouvée pour ce message.')],
        });
        return;
    }

    const memberRoles = interaction.member.roles;
    const hasOffRadioRole = memberRoles.cache.has(lsmsDutyData.offRadioRoleId);
    let successMessage = '';

    if (hasOffRadioRole) {
        await memberRoles.remove(lsmsDutyData.offRadioRoleId);
        successMessage = "Vous n'êtes plus off radio.";
    } else {
        await memberRoles.add(lsmsDutyData.offRadioRoleId);
        successMessage = 'Vous êtes maintenant off radio.';
    }

    await interaction.editReply({ embeds: [lsmsSuccessEmbedGenerator(successMessage)] });
}
