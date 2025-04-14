import { logger, player } from '@/index';
import { back } from '@/interactions/commands/music/back';
import { skip } from '@/interactions/commands/music/skip';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { generateNextMusicsWithGoogle } from '@/utils/intelligence';
import { waitTime } from '@/utils/utils';
import { QueryType, QueueRepeatMode, useQueue } from 'discord-player';
import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonInteraction,
    ButtonStyle,
    EmbedBuilder,
    GuildMember,
    MessageFlags,
} from 'discord.js';

export function backButton(interaction: ButtonInteraction) {
    back(interaction);
}

export async function loopButton(interaction: ButtonInteraction) {
    const methods = ['disabled', 'track', 'queue'];
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            flags: [MessageFlags.Ephemeral],
        });

    if (queue.repeatMode === QueueRepeatMode.QUEUE) {
        queue.setRepeatMode(QueueRepeatMode.OFF);
    } else {
        queue.setRepeatMode((queue.repeatMode + 1) as QueueRepeatMode);
    }

    return await interaction.reply({
        embeds: [successEmbed(interaction, `Boucle ${methods[queue.repeatMode]}`)],
        flags: [MessageFlags.Ephemeral],
    });
}

export async function resumeAndPauseButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            flags: [MessageFlags.Ephemeral],
        });

    const resumed = queue.node.resume();

    if (!resumed) {
        queue.node.pause();
        await interaction.reply({ embeds: [successEmbed(interaction, 'Musique mise en pause')] });
        await waitTime(5000);
        await interaction.deleteReply();
        return;
    }

    await interaction.reply({ embeds: [successEmbed(interaction, 'Musique reprise')] });
    await waitTime(5000);
    await interaction.deleteReply();
}

export function skipButton(interaction: ButtonInteraction) {
    skip(interaction);
}

export async function iaButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            flags: [MessageFlags.Ephemeral],
        });

    const track = queue.currentTrack;
    if (!track)
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune musique trouvée.'))],
            flags: [MessageFlags.Ephemeral],
        });
    const generatedMusics = await generateNextMusicsWithGoogle(`${track.title} - ${track.author}`);
    if (!generatedMusics)
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune musique trouvée.'))],
            flags: [MessageFlags.Ephemeral],
        });
    const embed = new EmbedBuilder()
        .setTitle('Suggestions de musique')
        .setDescription(`Voici quelques suggestions de musique après **${track.title}**`)
        .setColor('Green')
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user?.displayAvatarURL() })
        .setTimestamp();

    generatedMusics.forEach((music, index) => {
        embed.addFields({
            name: `Suggestion ${index + 1}`,
            value: music,
        });
    });

    const buttons = [
        new ButtonBuilder()
            .setCustomId('addIaSuggestion--1')
            .setLabel('Ajouter la 1ère suggestion')
            .setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('addIaSuggestion--2')
            .setLabel('Ajouter la 2ème suggestion')
            .setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('addIaSuggestion--3')
            .setLabel('Ajouter la 3ème suggestion')
            .setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('addIaSuggestion--4')
            .setLabel('Ajouter la 4ème suggestion')
            .setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('addIaSuggestion--5')
            .setLabel('Ajouter la 5ème suggestion')
            .setStyle(ButtonStyle.Primary),
    ];

    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    return interaction.reply({ embeds: [embed], flags: [MessageFlags.Ephemeral], components: [actionRow] });
}

export async function addIaSuggestion(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string);

    if (!queue?.isPlaying())
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))],
            flags: [MessageFlags.Ephemeral],
        });

    const fromEmbed = interaction.message.embeds[0];
    if (!fromEmbed)
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune musique trouvée.'))],
            flags: [MessageFlags.Ephemeral],
        });

    const trackSelected = fromEmbed.fields[Number(interaction.customId.split('--')[1]) - 1].value;
    if (!trackSelected)
        return interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucune musique trouvée.'))],
            flags: [MessageFlags.Ephemeral],
        });

    const userVoiceChannel = (interaction.member as GuildMember)?.voice.channel;
    if (!userVoiceChannel) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Vous devez être dans un salon vocal.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    const res = await player.search(trackSelected, {
        requestedBy: interaction.user,
        searchEngine: QueryType.AUTO,
    });

    if (!res?.tracks.length) {
        return await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Aucun résultat trouvé.'))],
            flags: [MessageFlags.Ephemeral],
        });
    }

    try {
        const { track } = await player.play(userVoiceChannel, trackSelected, {
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
            flags: [MessageFlags.Ephemeral],
        });
    }
}
