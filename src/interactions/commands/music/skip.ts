import { config } from '@/config';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { useQueue } from 'discord-player';
import { ButtonInteraction, CommandInteraction, MessageFlags, SlashCommandBuilder } from 'discord.js';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('skip')
    .setDescription('[Musique] Passer à la musique suivante');

export async function execute(interaction: CommandInteraction) {
    skip(interaction);
}

export async function skip(interaction: CommandInteraction | ButtonInteraction) {
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

    await queue.node.skip();

    await interaction.reply({ embeds: [successEmbed(interaction, 'Musique suivante')] });
    await waitTime(5000);
    await interaction.deleteReply();
}
