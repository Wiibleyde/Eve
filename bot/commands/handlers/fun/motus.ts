import { MessageFlags, SlashCommandBuilder, TextChannel } from 'discord.js';
import type { ICommand } from '../../command';
import { getRandomWord, MotusGame, motusGames } from '../../../../utils/games/motus';
import { logger } from '../../../..';
import { errorEmbedGenerator, successEmbedGenerator } from '../../../utils/embeds';

export const motus: ICommand = {
    data: new SlashCommandBuilder().setName('motus').setDescription('Lance un motus'),
    execute: async (interaction) => {
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });
        const word = await getRandomWord();
        logger.debug('Motus word: ', word);
        const game = new MotusGame(word, interaction.user.id);

        const embed = await game.getEmbed();

        const channel = interaction.channel as TextChannel;
        if (!channel) {
            logger.warn('Impossible de trouver le canal');
            const errorEmbed = errorEmbedGenerator('Impossible de trouver le canal');
            await interaction.editReply({ embeds: [errorEmbed] });
            return;
        }

        const message = await channel.send({
            embeds: [embed.embed],
            components: embed.components,
        });

        motusGames.set(message.id, game);
        const successEmbed = successEmbedGenerator('Partie de motus lancée !');
        await interaction.editReply({ embeds: [successEmbed] });
    },
};
