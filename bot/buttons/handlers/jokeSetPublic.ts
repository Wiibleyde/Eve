import { MessageFlags, type ButtonInteraction } from 'discord.js';
import {
    jokeBasicEmbedGenerator,
    jokeErrorEmbedGenerator,
    jokeSuccessEmbedGenerator,
} from '../../commands/handlers/blague';

export async function jokeSetPublicButton(interaction: ButtonInteraction) {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const messageEmbed = interaction.message.embeds[0];
    if (!messageEmbed) {
        const errorEmbed = jokeErrorEmbedGenerator('Aucune blague trouvée');
        await interaction.editReply({ embeds: [errorEmbed] });
        return;
    }
    const joke = messageEmbed.description;
    const answer = messageEmbed.fields[0]?.value;
    if (!joke || !answer) {
        const errorEmbed = jokeErrorEmbedGenerator('Aucune blague trouvée');
        await interaction.editReply({ embeds: [errorEmbed] });
        return;
    }

    const publicEmbed = jokeBasicEmbedGenerator();
    publicEmbed.setDescription(joke);
    publicEmbed.addFields({
        name: 'Réponse :',
        value: answer,
    });
    publicEmbed.addFields({
        name: 'Rendu publique par :',
        value: `<@${interaction.user.id}>`,
    });

    const channel = interaction.client.channels.cache.get(interaction.channelId);
    if (!channel || !channel.isTextBased() || !('send' in channel) || typeof channel.send !== 'function') {
        const errorEmbed = jokeErrorEmbedGenerator("Impossible d'envoyer la blague dans ce canal.");
        await interaction.editReply({ embeds: [errorEmbed] });
        return;
    }

    await channel.send({ embeds: [publicEmbed] });

    const successEmbed = jokeSuccessEmbedGenerator('Blague envoyée !');
    await interaction.editReply({ embeds: [successEmbed] });
}
