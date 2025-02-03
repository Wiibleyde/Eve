import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { useQueue } from 'discord-player';
import { CommandInteraction, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('remove')
    .setDescription("[Musique] Supprimer une musique de la file d'attente")
    .addStringOption((option) =>
        option.setName('position').setDescription('La position de la musique à supprimer').setRequired(true)
    );

export async function execute(interaction: CommandInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const position = interaction.options.get('position')?.value as string;

    if (isNaN(Number(position)))
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('La position doit être un nombre.'))],
            ephemeral: true,
        });

    const index = Number(position) - 1;
    const name = queue.tracks.toArray()[index]?.title;
    if (!name)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Cette musique n'existe pas."))],
            ephemeral: true,
        });

    queue.removeTrack(index);

    await interaction.reply({
        embeds: [successEmbed(interaction, `La musique ${name} a été supprimée de la file d'attente.`)],
    });
    await waitTime(5000);
    await interaction.deleteReply();
}
