import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    MessageFlags,
    UserContextMenuCommandInteraction,
} from 'discord.js';
import type { IContextMenuUserCommand } from '../contextMenuUser';
import { basicEmbedGenerator } from '../../../utils/embeds';

export const profilePicture: IContextMenuUserCommand = {
    data: new ContextMenuCommandBuilder().setName('Récupèrer la photo de profil').setType(ApplicationCommandType.User),
    async execute(interaction: UserContextMenuCommandInteraction) {
        await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });

        const author = interaction.targetUser;
        const userProfilePicture = author?.displayAvatarURL({ extension: 'png', size: 1024 });

        const embed = basicEmbedGenerator().setTitle('Photo de profil').setImage(userProfilePicture);

        await interaction.editReply({ embeds: [embed] });
    },
};
