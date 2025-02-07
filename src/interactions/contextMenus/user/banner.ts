import { errorEmbed } from '@/utils/embeds';
import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    EmbedBuilder,
    MessageFlags,
    UserContextMenuCommandInteraction,
} from 'discord.js';

export const data: ContextMenuCommandBuilder = new ContextMenuCommandBuilder()
    .setName('Récupèrer la bannière')
    .setType(ApplicationCommandType.User);

export async function execute(interaction: UserContextMenuCommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });

    const author = interaction.targetUser;

    let userBanner = author?.bannerURL({ size: 1024, extension: 'png' });

    if (!userBanner) {
        try {
            const user = await interaction.client.users.fetch(author.id, { force: true });
            userBanner = user.bannerURL({ size: 1024, extension: 'png' });
        } catch {
            await interaction.editReply({
                embeds: [
                    errorEmbed(interaction, new Error("Impossible de récupérer les informations de l'utilisateur.")),
                ],
            });
            return;
        }
    }

    if (!userBanner) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("L'utilisateur n'a pas de bannière."))],
        });
        return;
    }

    const embed = new EmbedBuilder()
        .setTitle('Bannière')
        .setImage(userBanner)
        .setColor(0x00ff00)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    await interaction.editReply({ embeds: [embed] });
}
