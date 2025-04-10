import { errorEmbed, successEmbed } from '@/utils/embeds';
import { TicTacToeGame, tictactoeGames } from '@/utils/games/tictactoe';
import {
    CommandInteraction,
    MessageFlags,
    MessagePayload,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
} from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('tictactoe')
    .setDescription('Lance une partie de Tic Tac Toe.')
    .addUserOption((option) => option.setName('opponent').setDescription("L'opposant à défier").setRequired(true));

export async function execute(interaction: CommandInteraction) {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const opponent = interaction.options.get('opponent')?.user;
    if (!opponent) {
        return await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error("Impossible de trouver l'opposant."))],
            flags: [MessageFlags.Ephemeral],
        });
    }
    if (opponent.id === interaction.user.id) {
        return await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error('Vous ne pouvez pas jouer contre vous-même.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
    if (opponent.bot) {
        return await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error('Vous ne pouvez pas jouer contre un bot.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const channel = interaction.channel;
    if (!channel) {
        return await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error('Impossible de trouver le salon de jeu.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const player1 = interaction.user.id;
    const player2 = opponent.id;
    const game = new TicTacToeGame(player1, player2);

    const response = (await game.getResponse()) as MessagePayload;
    if (channel.isSendable()) {
        const message = await channel.send(response);
        tictactoeGames.set(message.id, game);
        await interaction.followUp({
            embeds: [successEmbed(interaction, 'Partie de Tic Tac Toe lancée. C’est au tour de X de jouer.')],
            flags: [MessageFlags.Ephemeral],
        });
    } else {
        await interaction.followUp({
            embeds: [errorEmbed(interaction, new Error("Impossible d'envoyer le message dans le salon."))],
            flags: [MessageFlags.Ephemeral],
        });
    }
}
