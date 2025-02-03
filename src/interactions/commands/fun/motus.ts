import { errorEmbed, successEmbed } from '@/utils/embeds';
import { games, getRandomWord, MotusGame } from '@/utils/games/motus';
import { CommandInteraction, SlashCommandBuilder, TextChannel } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('motus')
    .setDescription('Lance une partie de Motus.');

export async function execute(interaction: CommandInteraction) {
    const word = await getRandomWord();
    const game = new MotusGame(word, interaction.user.id);

    const embedResult = await game.getEmbed();

    const channel = interaction.channel as TextChannel;
    if (!channel) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de trouver le salon de jeu.'))],
            ephemeral: true,
        });
    }

    const message = await channel.send({
        embeds: [embedResult.embed],
        components: embedResult.components,
        files: embedResult.attachments,
    });

    games.set(message.id, game);

    await interaction.reply({ embeds: [successEmbed(interaction, 'Partie de Motus lancée.')], ephemeral: true });
}
