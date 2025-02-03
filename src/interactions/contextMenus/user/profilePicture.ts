import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    EmbedBuilder,
    UserContextMenuCommandInteraction,
} from 'discord.js';

export const data: ContextMenuCommandBuilder = new ContextMenuCommandBuilder()
    .setName('Récupèrer la photo de profil')
    //@ts-expect-error - This is a valid type
    .setType(ApplicationCommandType.User);

export async function execute(interaction: UserContextMenuCommandInteraction): Promise<void> {
    await interaction.deferReply({ ephemeral: true });

    const author = interaction.targetUser;
    const userProfilePicture = author?.displayAvatarURL({ extension: 'png', size: 1024 });

    const embed = new EmbedBuilder()
        .setTitle('Photo de profil')
        .setImage(userProfilePicture)
        .setColor(0x00ff00)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    await interaction.editReply({ embeds: [embed] });
}
