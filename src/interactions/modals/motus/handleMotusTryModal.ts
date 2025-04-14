import { client } from '@/index';
import { errorEmbed } from '@/utils/embeds';
import { games, GameState, TryReturn } from '@/utils/games/motus';
import { MessageFlags, ModalSubmitInteraction, TextChannel } from 'discord.js';

export async function handleMotusTryModal(interaction: ModalSubmitInteraction) {
    const messageId = interaction.customId.split('--')[1];
    const game = games.get(messageId);

    if (!interaction.channel) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de trouver le salon de jeu.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }
    const message = await client.channels
        .fetch(interaction.channel.id)
        .then((channel) => (channel as TextChannel).messages.fetch(messageId));

    if (!game) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune partie de Motus en cours.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    if (game.state !== GameState.PLAYING) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('La partie de Motus est déjà terminée.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const wordLength = game.wordLength;
    const word = interaction.fields.getTextInputValue('motusTry');

    if (word.length !== wordLength) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error(`Le mot doit contenir ${wordLength} lettres.`))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const result = game.tryAttempt(word, interaction.user.id);

    if (result === TryReturn.WIN) {
        game.endGame(GameState.WON);
    } else if (result === TryReturn.LOSE) {
        game.endGame(GameState.LOST);
    }

    const embedResult = await game.getEmbed();

    if (embedResult.components.length > 0) {
        await message.edit({ embeds: [embedResult.embed], components: embedResult.components });
    } else {
        await message.edit({ embeds: [embedResult.embed], components: [] });
    }

    if (game.state !== GameState.PLAYING) {
        games.delete(message.id);
    }

    await interaction.deferUpdate();
}
