import { config } from '@/config';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { QueueRepeatMode, useQueue } from 'discord-player';
import { CommandInteraction, MessageFlags, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('loop')
    .setDescription('[Musique] Activer/Désactiver la boucle')
    .addStringOption((option) =>
        option
            .setName('action')
            .setDescription('Mode de boucle')
            .addChoices([
                { name: 'Queue', value: 'enable_loop_queue' },
                { name: 'Disable', value: 'disable_loop' },
                { name: 'Song', value: 'enable_loop_song' },
                { name: 'Autoplay', value: 'enable_loop_autoplay' },
            ])
            .setRequired(true)
    );

export async function execute(interaction: CommandInteraction) {
    if (config.MUSIC_MODULE !== true) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Module de musique désactivé.'))],
            ephemeral: true,
        });
        return;
    }
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            ephemeral: true,
        });

    switch (interaction.options.get('action')?.value as string) {
        case 'enable_loop_queue': {
            if (queue.repeatMode === QueueRepeatMode.TRACK)
                return await interaction.reply({
                    embeds: [
                        errorEmbed(interaction, new Error("Impossible d'activer la boucle de la file d'attente.")),
                    ],
                    ephemeral: true,
                });

            queue.setRepeatMode(QueueRepeatMode.QUEUE);

            await interaction.reply({
                embeds: [successEmbed(interaction, "Boucle de la file d'attente activée")],
                ephemeral: true,
            });
            break;
        }
        case 'disable_loop': {
            if (queue.repeatMode === QueueRepeatMode.OFF)
                return await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error('Impossible de désactiver la boucle.'))],
                    ephemeral: true,
                });

            queue.setRepeatMode(QueueRepeatMode.OFF);

            await interaction.reply({ embeds: [successEmbed(interaction, 'Boucle désactivée')], flags: [MessageFlags.Ephemeral] });
            break;
        }
        case 'enable_loop_song': {
            if (queue.repeatMode === QueueRepeatMode.QUEUE)
                return await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error("Impossible d'activer la boucle de la musique."))],
                    ephemeral: true,
                });

            queue.setRepeatMode(QueueRepeatMode.TRACK);

            await interaction.reply({
                embeds: [successEmbed(interaction, 'Boucle de la musique activée')],
                ephemeral: true,
            });
            break;
        }
        case 'enable_loop_autoplay': {
            if (queue.repeatMode === QueueRepeatMode.TRACK)
                return await interaction.reply({
                    embeds: [errorEmbed(interaction, new Error("Impossible d'activer la boucle de l'autoplay."))],
                    ephemeral: true,
                });

            queue.setRepeatMode(QueueRepeatMode.AUTOPLAY);

            await interaction.reply({
                embeds: [successEmbed(interaction, "Boucle de l'autoplay activée")],
                ephemeral: true,
            });
            break;
        }
        default: {
            await interaction.reply({
                embeds: [errorEmbed(interaction, new Error('Action inconnue.'))],
                ephemeral: true,
            });
            break;
        }
    }
}
