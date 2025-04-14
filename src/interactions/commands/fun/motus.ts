import { logger } from '@/index';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { games, getRandomWord, MotusGame } from '@/utils/games/motus';
import { CommandInteraction, MessageFlags, SlashCommandBuilder, TextChannel } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('motus')
    .setDescription('Lance une partie de Motus.');

export async function execute(interaction: CommandInteraction) {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const word = await getRandomWord();
    const game = new MotusGame(word, interaction.user.id);

    const embedResult = await game.getEmbed();

    const channel = interaction.channel as TextChannel;
    if (!channel) {
        logger.warn('Impossible de trouver le salon de jeu.');
        return await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error('Impossible de trouver le salon de jeu.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const message = await channel.send({
        embeds: [embedResult.embed],
        components: embedResult.components,
        files: embedResult.attachments,
    });

    games.set(message.id, game);

    await interaction.followUp({
        embeds: [successEmbed(interaction, 'Partie de Motus lancée.')],
        flags: [MessageFlags.Ephemeral],
    });
}
