import { MessageFlags, ModalSubmitInteraction, TextChannel } from 'discord.js';
import { GameState, motusGames, TryReturn } from '../../../utils/games/motus';
import { errorEmbedGenerator } from '../../utils/embeds';
import { client } from '../../bot';

export async function handleMotusTryModal(interaction: ModalSubmitInteraction): Promise<void> {
    // Early validation for required objects
    if (!interaction.message || !interaction.channel) {
        await replyWithError(interaction, 'Impossible de trouver le message ou le salon de jeu.');
        return;
    }

    const messageId = interaction.message.id;

    // Try to fetch the original game message
    const message = await fetchGameMessage(interaction.channel.id, messageId);
    if (!message) {
        await replyWithError(interaction, 'Impossible de trouver le message de jeu.');
        return;
    }

    // Get the game from the map
    const game = motusGames.get(messageId);
    if (!game) {
        await replyWithError(interaction, 'Aucune partie de Motus en cours.');
        return;
    }

    // Check if the game is still active
    if (game.state !== GameState.PLAYING) {
        await replyWithError(interaction, 'La partie de Motus est déjà terminée.');
        return;
    }

    // Get and validate the word input
    const wordLength = game.wordLength;
    const word = interaction.fields.getTextInputValue('motusTry');
    if (word.length !== wordLength) {
        await replyWithError(interaction, `Le mot doit contenir ${wordLength} lettres.`);
        return;
    }

    // Process the attempt
    const result = game.tryAttempt(word, interaction.user.id);

    // Check game result
    if (result === TryReturn.WIN) {
        game.endGame(GameState.WON);
    } else if (result === TryReturn.LOSE) {
        game.endGame(GameState.LOST);
    }

    // Update the game embed
    const embedResult = await game.getEmbed();
    await message.edit({
        embeds: [embedResult.embed],
        components: embedResult.components.length > 0 ? embedResult.components : [],
    });

    // Clean up if game is finished
    if (game.state !== GameState.PLAYING) {
        motusGames.delete(messageId);
    }

    await interaction.deferUpdate();
}

/**
 * Helper function to reply with an error message
 */
async function replyWithError(interaction: ModalSubmitInteraction, errorMessage: string) {
    return await interaction.reply({
        embeds: [errorEmbedGenerator(errorMessage)],
        flags: [MessageFlags.Ephemeral],
    });
}

/**
 * Helper function to fetch the game message
 */
async function fetchGameMessage(channelId: string, messageId: string) {
    try {
        const channel = await client.channels.fetch(channelId);
        if (!channel || !(channel instanceof TextChannel)) return null;
        return await channel.messages.fetch(messageId);
    } catch {
        return null;
    }
}
