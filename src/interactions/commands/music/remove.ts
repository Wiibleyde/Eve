import { config } from '@/config';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { useQueue } from 'discord-player';
import { CommandInteraction, MessageFlags, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('remove')
    .setDescription("[Musique] Supprimer une musique de la file d'attente")
    .addStringOption((option) =>
        option.setName('position').setDescription('La position de la musique à supprimer').setRequired(true)
    );

export async function execute(interaction: CommandInteraction) {
    if (config.MUSIC_MODULE !== true) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Module de musique désactivé.'))],
            flags: [MessageFlags.Ephemeral],
        });
        return;
    }
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            flags: [MessageFlags.Ephemeral],
        });

    const position = interaction.options.get('position')?.value as string;

    if (isNaN(Number(position)))
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('La position doit être un nombre.'))],
            flags: [MessageFlags.Ephemeral],
        });

    const index = Number(position) - 1;
    const name = queue.tracks.toArray()[index]?.title;
    if (!name)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Cette musique n'existe pas."))],
            flags: [MessageFlags.Ephemeral],
        });

    queue.removeTrack(index);

    await interaction.reply({
        embeds: [successEmbed(interaction, `La musique ${name} a été supprimée de la file d'attente.`)],
    });
    await waitTime(5000);
    await interaction.deleteReply();
}
