import { laboInQueryManager, type LaboInQueryEntry } from '@utils/rp/labo';
import { lsmsEmbedGenerator, lsmsErrorEmbedGenerator, lsmsSuccessEmbedGenerator } from '@utils/rp/lsms';
import { MessageFlags, type ButtonInteraction } from 'discord.js';

export async function laboCancelButton(interaction: ButtonInteraction): Promise<void> {
    if (!interaction.guildId || !interaction.inCachedGuild()) {
        await interaction.reply({
            embeds: [lsmsErrorEmbedGenerator('Cette commande ne peut être utilisée que dans un serveur Discord.')],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    const messageId = interaction.message.id;
    const { success: cancelled, entry } = laboInQueryManager.cancelByMessageId(messageId);
    if (cancelled && entry) {
        const cancelEmbed = lsmsEmbedGenerator()
            .setTitle('Analyse en cours')
            .setDescription(`Analyse pour **${entry.name}** annulée par <@${interaction.user.id}>.`)
            .addFields(
                {
                    name: "Type d'analyse",
                    value: laboInQueryManager.getAnalyseType({ type: entry.type } as LaboInQueryEntry),
                    inline: true,
                },
                { name: 'Nom de la personne', value: entry.name, inline: true },
                { name: 'Demandé par', value: `<@${interaction.user.id}>`, inline: true }
            );
        await interaction.message.edit({ embeds: [cancelEmbed], components: [] });
        await interaction.editReply({
            embeds: [lsmsSuccessEmbedGenerator("L'analyse a été annulée avec succès.")],
        });
    } else {
        await interaction.editReply({
            embeds: [lsmsErrorEmbedGenerator('Aucune analyse en cours trouvée pour ce message.')],
        });
    }
}
