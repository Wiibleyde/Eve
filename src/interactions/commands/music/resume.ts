import { config } from '@/config';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { useQueue } from 'discord-player';
import { CommandInteraction, MessageFlags, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('resume')
    .setDescription('[Musique] Reprendre la musique');

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

    if (queue.node.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de reprendre la musique.'))],
            flags: [MessageFlags.Ephemeral],
        });

    const success = queue.node.resume();

    if (!success)
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de mettre en pause la musique.'))],
            flags: [MessageFlags.Ephemeral],
        });

    await interaction.reply({ embeds: [successEmbed(interaction, 'Musique mise en pause')] });
    await waitTime(5000);
    await interaction.deleteReply();
}
