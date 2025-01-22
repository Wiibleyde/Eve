import { playerConfig } from "@/config";
import { client, logger, player } from "..";
import { ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder, MessageActionRowComponentBuilder } from "discord.js";

player.events.on("audioTrackAdd", (queue, track) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Musique ajoutée")
            .setDescription(`[${track.title}](${track.url}) a été ajoutée à la file d'attente.`)
            .setThumbnail(track.thumbnail)
            .setColor('Green')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
});

player.events.on("audioTracksAdd", (queue, tracks) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Musiques ajoutées")
            .setDescription(`${tracks.length} musiques ont été ajoutées à la file d'attente.`)
            .setColor('Green')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
});

player.events.on("disconnect", (queue) => {
    if(queue.metadata.lyricsThread) {
        queue.metadata.lyricsThread.delete()
        queue.setMetadata({channel: queue.metadata.channel})
    }

    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Déconnexion")
            .setDescription(`Déconnexion du salon vocal.`)
            .setColor('Red')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
})

player.events.on("emptyChannel", (queue) => {
    if (queue.metadata.lyricsThread) {
        queue.metadata.lyricsThread.delete()
        queue.setMetadata({ channel: queue.metadata.channel })
    }

    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("File d'attente vide")
            .setDescription(`La file d'attente est vide, je quitte le salon vocal.`)
            .setColor('Red')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
})

player.events.on("emptyQueue", (queue) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("File d'attente vide")
            .setDescription(`La file d'attente est vide.`)
            .setColor('Red')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
})

player.events.on("error", (queue, error) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Erreur")
            .setDescription(`Une erreur est survenue lors de la lecture de la musique : ${error.message}`)
            .setColor('Red')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }

    logger.error(`An error occurred while playing music: ${error.message}`)
})

player.events.on("playerError", (queue, error) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Erreur")
            .setDescription(`Une erreur est survenue lors de la lecture de la musique : ${error.message}`)
            .setColor('Red')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }

    logger.error(`An error occurred while playing music: ${error.message}`)
})

player.events.on("playerFinish", (queue) => {
    if (queue.metadata.lyricsThread) {
        queue.metadata.lyricsThread.delete();
        queue.setMetadata({
            channel: queue.metadata.channel
        })
    }
})

player.events.on("playerSkip", (queue, track) => {
    if(playerConfig.extraMessages) {
        const embed = new EmbedBuilder()
            .setTitle("Musique sautée")
            .setDescription(`[${track.title}](${track.url}) a été sautée.`)
            .setThumbnail(track.thumbnail)
            .setColor('Yellow')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
            .setTimestamp()

        queue.metadata.channel.send({ embeds: [embed] })
    }
})

player.events.on("playerStart", async (queue, track) => {
    if (!playerConfig.loopMessage && queue.repeatMode !== 0) return

    let emojiState = playerConfig.enableEmoji

    const emojis = playerConfig.emojis

    emojiState = emojis ? true : false

    const embed = new EmbedBuilder()
        .setTitle("Lecture en cours")
        .setDescription(`[${track.title}](${track.url}) a commencé à jouer dans <#${queue.channel?.id}>.`)
        .setThumbnail(track.thumbnail)
        .setColor('Green')
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() })
        .setTimestamp()

    const back = new ButtonBuilder()
        .setLabel(emojiState ? emojis.back : ('Back'))
        .setCustomId('backButton')
        .setStyle(ButtonStyle.Primary)

    const skip = new ButtonBuilder()
        .setLabel(emojiState ? emojis.skip : ('Skip'))
        .setCustomId('skipButton')
        .setStyle(ButtonStyle.Primary)

    const resumepause = new ButtonBuilder()
        .setLabel(emojiState ? emojis.ResumePause : ('Resume/Pause'))
        .setCustomId('resumeAndPauseButton')
        .setStyle(ButtonStyle.Danger)

    const loop = new ButtonBuilder()
        .setLabel(emojiState ? emojis.loop : ('Loop'))
        .setCustomId('loopButton')
        .setStyle(ButtonStyle.Danger)

    const row = new ActionRowBuilder<MessageActionRowComponentBuilder>().addComponents(back, skip, resumepause, loop)

    if (queue.metadata.playingMessage) {
        try {
            await queue.metadata.playingMessage.edit({ embeds: [embed], components: [row] })
        } catch {
            queue.metadata.playingMessage = await queue.metadata.channel.send({ embeds: [embed], components: [row] })
        }
    } else {
        queue.metadata.playingMessage = await queue.metadata.channel.send({ embeds: [embed], components: [row] })
    }
})