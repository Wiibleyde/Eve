import { errorEmbed, successEmbed } from "@/utils/embeds";
import { tictactoeGames, TicTacToeGameState } from "@/utils/games/tictactoe";
import { ButtonInteraction, CacheType, InteractionUpdateOptions, MessageFlags } from "discord.js";

export async function handleTicTacToeButton(interaction: ButtonInteraction<CacheType>): Promise<void> {
    const message = interaction.message;
    if (!message) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error(
                        'Une erreur est survenue lors de la récupération de la question de quiz (message introuvable).'
                    )
                ),
            ],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }
    const game = tictactoeGames.get(message.id);
    if (!game) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error(
                        'Une erreur est survenue lors de la récupération de la question de quiz (message introuvable).'
                    )
                ),
            ],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const casePlayed = parseInt(interaction.customId.split('--')[1], 10);
    const row = Math.floor(casePlayed / 3);
    const col = casePlayed % 3;

    if (!game.playMove(row, col, interaction.user.id)) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error('Ce coup est invalide ou le jeu est déjà terminé, ou ce n’est pas votre tour.')
                ),
            ],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }

    const updatedResponse = await game.getResponse() as InteractionUpdateOptions;
    await interaction.update(updatedResponse as InteractionUpdateOptions);

    if (game.getGameState() !== TicTacToeGameState.IN_PROGRESS) {
        tictactoeGames.delete(message.id);
        await interaction.followUp({
            embeds: [
                successEmbed(
                    interaction,
                    'Le jeu est terminé !'
                ),
            ],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }
    return;
}