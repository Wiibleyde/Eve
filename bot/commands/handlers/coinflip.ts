import { MessageFlags, SlashCommandBuilder } from 'discord.js';
import type { ICommand } from '../command';
import { basicEmbedGenerator } from '@bot/utils/embeds';

export const coinflip: ICommand = {
    data: new SlashCommandBuilder().setName('coinflip').setDescription('Lance une pièce et affiche le résultat.'),
    execute: async (interaction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        const result = Math.random() < 0.5 ? 'Pile' : 'Face';
        await interaction.editReply({
            embeds: [basicEmbedGenerator().setDescription(`Le résultat du lancer de pièce est : **${result}**.`)],
        });
    },
};
