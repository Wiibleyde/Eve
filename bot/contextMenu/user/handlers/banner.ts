import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    MessageFlags,
    UserContextMenuCommandInteraction,
} from 'discord.js';
import type { IContextMenuUserCommand } from '../contextMenuUser';
import { basicEmbedGenerator, errorEmbedGenerator } from '../../../utils/embeds';

export const banner: IContextMenuUserCommand = {
    data: new ContextMenuCommandBuilder().setName('Récupèrer la bannière').setType(ApplicationCommandType.User),
    async execute(interaction: UserContextMenuCommandInteraction) {
        await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });

        const author = interaction.targetUser;

        let userBanner = author?.bannerURL({ size: 1024, extension: 'png' });

        if (!userBanner) {
            try {
                const user = await interaction.client.users.fetch(author.id, { force: true });
                userBanner = user.bannerURL({ size: 1024, extension: 'png' });
            } catch {
                await interaction.editReply({
                    embeds: [errorEmbedGenerator("Impossible de récupérer les informations de l'utilisateur.")],
                });
                return;
            }
        }

        if (!userBanner) {
            await interaction.editReply({
                embeds: [errorEmbedGenerator("L'utilisateur n'a pas de bannière.")],
            });
            return;
        }

        const embed = basicEmbedGenerator().setTitle('Bannière').setImage(userBanner);

        await interaction.editReply({ embeds: [embed] });
    },
};
