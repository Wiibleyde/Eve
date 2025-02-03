import { errorEmbed, successEmbed } from '@/utils/embeds';
import { capitalizeFirstLetter } from '@/utils/textUtils';
import { AudioFilters, QueueFilters, useQueue } from 'discord-player';
import { CommandInteraction, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('filter')
    .setDescription('[Musique] Activer/Désactiver un filtre')
    .addStringOption((option) =>
        option
            .setName('filter')
            .setDescription('Nom du filtre')
            .addChoices([
                ...Object.keys(AudioFilters.filters)
                    .map((filter) => Object({ name: capitalizeFirstLetter(filter), value: filter }))
                    .splice(0, 25),
            ])
            .setRequired(true)
    );

export async function execute(interaction: CommandInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    const selectedFilter = interaction.options.get('filter')?.value as string;

    const filters: string[] = [];
    queue.filters.ffmpeg.getFiltersDisabled().forEach((filter) => filters.push(filter));
    queue.filters.ffmpeg.getFiltersEnabled().forEach((filter) => filters.push(filter));

    const filter = filters.find((f) => f.toLowerCase() === selectedFilter.toLowerCase()) as keyof QueueFilters;

    if (!filter)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Filtre non trouvé.'))],
            ephemeral: true,
        });

    await queue.filters.ffmpeg.toggle(filter);

    await interaction.reply({
        embeds: [
            successEmbed(
                interaction,
                `${capitalizeFirstLetter(filter)} ${queue.filters.ffmpeg.getFiltersEnabled().includes(filter) ? 'activé' : 'désactivé'}`
            ),
        ],
        ephemeral: true,
    });
}
