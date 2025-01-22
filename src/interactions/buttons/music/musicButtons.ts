import { back } from "@/interactions/commands/music/back";
import { skip } from "@/interactions/commands/music/skip";
import { errorEmbed, successEmbed } from "@/utils/embeds";
import { waitTime } from "@/utils/utils";
import { QueueRepeatMode, useQueue } from "discord-player";
import { ButtonInteraction } from "discord.js";

export function backButton(interaction: ButtonInteraction) {
    back(interaction)
}

export async function loopButton(interaction: ButtonInteraction) {
    const methods = ['disabled', 'track', 'queue']
    const queue = useQueue(interaction.guildId as string)

    if (!queue?.isPlaying()) return await interaction.reply({ embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))], ephemeral: true })

    if(queue.repeatMode === QueueRepeatMode.QUEUE) {
        queue.setRepeatMode(QueueRepeatMode.OFF)
    } else {
        queue.setRepeatMode(queue.repeatMode + 1)
    }

    return await interaction.reply({ embeds: [successEmbed(interaction, `Boucle ${methods[queue.repeatMode]}`)], ephemeral: true })
}

export async function resumeAndPauseButton(interaction: ButtonInteraction) {
    const queue = useQueue(interaction.guildId as string)

    if(!queue?.isPlaying()) return await interaction.reply({ embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))], ephemeral: true })

    const resumed = queue.node.resume()

    if(!resumed) {
        queue.node.pause()
        await interaction.reply({ embeds: [successEmbed(interaction, "Musique mise en pause")] })
        await waitTime(5000)
        await interaction.deleteReply()
        return
    }

    await interaction.reply({ embeds: [successEmbed(interaction, "Musique reprise")] })
    await waitTime(5000)
    await interaction.deleteReply()
}

export function skipButton(interaction: ButtonInteraction) {
    skip(interaction)
}
