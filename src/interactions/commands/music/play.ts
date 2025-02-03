import { logger } from '@/index';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { waitTime } from '@/utils/utils';
import { QueryType, useMainPlayer } from 'discord-player';
import { CommandInteraction, GuildMember, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('play')
    .setDescription('[Musique] Jouer une musique')
    .addStringOption((option) => option.setName('musique').setDescription('Nom de la musique').setRequired(true));

export async function execute(interaction: CommandInteraction) {
    const player = useMainPlayer();

    const song = interaction.options.get('musique')?.value as string;

    const userVoiceChannel = (interaction.member as GuildMember)?.voice.channel;
    if (!userVoiceChannel) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Vous devez être dans un salon vocal.'))],
            ephemeral: true,
        });
    }

    const res = await player.search(song, {
        requestedBy: interaction.user,
        searchEngine: QueryType.AUTO,
    });

    if (!res?.tracks.length) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucun résultat trouvé.'))],
            ephemeral: true,
        });
    }

    try {
        const { track } = await player.play(userVoiceChannel, song, {
            nodeOptions: {
                metadata: {
                    channel: interaction.channel,
                },
                volume: 100,
                leaveOnEmpty: true,
                leaveOnEmptyCooldown: 60000,
                leaveOnEnd: true,
                leaveOnEndCooldown: 60000,
            },
        });

        await interaction.reply({
            embeds: [successEmbed(interaction, `Musique ajoutée à la file d'attente: [${track.title}](${track.url})`)],
        });
        await waitTime(5000);
        await interaction.deleteReply();
    } catch (error) {
        logger.error(error);
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Impossible de jouer la musique.'))],
            ephemeral: true,
        });
    }
}
