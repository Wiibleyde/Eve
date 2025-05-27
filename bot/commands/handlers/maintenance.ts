import { MessageFlags, SlashCommandBuilder } from 'discord.js';
import type { ICommand } from '../command';
import { successEmbedGenerator, warningEmbedGenerator } from '../../utils/embeds';
import { toggleMaintenanceMode } from '../../../utils/core/maintenance';
import { config } from '../../../utils/core/config';

export const maintenance: ICommand = {
    data: new SlashCommandBuilder().setName('maintenance').setDescription('Mettre le bot en mode maintenance'),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (interaction.user.id !== config.OWNER_ID) {
            await interaction.editReply({
                embeds: [warningEmbedGenerator("Vous n'avez pas la permission de mettre le bot en mode maintenance.")],
            });
            return;
        }
        const isMaintenance = toggleMaintenanceMode();
        await interaction.editReply({
            embeds: [
                successEmbedGenerator(
                    `Le bot est maintenant en mode maintenance : ${isMaintenance ? 'activé' : 'désactivé'}.`
                ),
            ],
        });
    },
};
