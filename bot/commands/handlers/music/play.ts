import type { ICommand } from "@bot/commands/command";
import { musicErrorEmbedGenerator, musicSuccessEmbedGenerator } from "@utils/music";
import { QueryType, useMainPlayer } from "discord-player";
import { GuildMember, MessageFlags, SlashCommandBuilder } from "discord.js";

export const play: ICommand = {
    data: new SlashCommandBuilder()
        .setName('play')
        .setDescription('[Musique] Jouer une musique')
        .addStringOption((option) => option.setName('musique').setDescription('Nom de la musique').setRequired(true)),
    execute: async (interaction) => {
        const player = useMainPlayer();
        const song = interaction.options.get('musique')?.value as string;
        const userVoiceChannel = (interaction.member as GuildMember)?.voice.channel;
        if (!userVoiceChannel) {
            await interaction.reply({
                embeds: [
                    musicErrorEmbedGenerator('Vous devez être dans un salon vocal pour jouer de la musique.'),
                ],
                flags: [MessageFlags.Ephemeral],
            });
            return;
        }

        const res = await player.search(song, {
            requestedBy: interaction.user,
            searchEngine: QueryType.AUTO,
        });
        if (!res?.tracks.length) {
            await interaction.reply({
                embeds: [musicErrorEmbedGenerator('Aucune musique trouvée.')],
                flags: [MessageFlags.Ephemeral],
            });
            return;
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
                embeds: [musicSuccessEmbedGenerator(`La musique **${track.title}** a été ajoutée à la file d'attente.`)],
                flags: [MessageFlags.Ephemeral],
            });
        } catch (error) {
            console.error('Erreur lors de la lecture de la musique:', error);
            await interaction.reply({
                embeds: [musicErrorEmbedGenerator('Une erreur est survenue lors de la lecture de la musique.')],
                flags: [MessageFlags.Ephemeral],
            });
        }
    }
}

