import { errorEmbed, successEmbed } from "@/utils/embeds";
import { waitTime } from "@/utils/utils";
import { useQueue } from "discord-player";
import { CommandInteraction, SlashCommandBuilder } from "discord.js";


export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName("stop")
    .setDescription("[Musique] Arrêter la musique")

export async function execute(interaction: CommandInteraction) {
    const queue = useQueue(interaction.guildId as string)

    if (!queue?.isPlaying()) return await interaction.reply({ embeds: [errorEmbed(interaction, new Error("Aucune musique n'est en cours de lecture."))], ephemeral: true })

    queue.delete()

    await interaction.reply({ embeds: [successEmbed(interaction, "La musique a été arrêtée.")] })
    await waitTime(5000)
    await interaction.deleteReply()
}