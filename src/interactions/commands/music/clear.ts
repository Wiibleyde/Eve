import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { useQueue } from 'discord-player';
import { ButtonInteraction, CommandInteraction, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('clear')
    .setDescription("[Musique] Vider la file d'attente");

export async function execute(interaction: CommandInteraction) {
    clear(interaction);
}

async function clear(interaction: CommandInteraction | ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    if (!queue.tracks.toArray()[1])
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Il n'y a pas de musique dans la file d'attente."))],
            ephemeral: true,
        });

    queue.tracks.clear();

    await interaction.reply({ embeds: [successEmbed(interaction, "File d'attente vidée")] });
    await waitTime(5000);
    await interaction.deleteReply();
}
