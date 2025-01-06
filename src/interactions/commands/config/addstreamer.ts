import { streamRetriever } from "@/index";
import { prisma } from "@/utils/database";
import { errorEmbed } from "@/utils/embeds";
import { hasPermission } from "@/utils/permissionTester";
import { CommandInteraction, PermissionFlagsBits, Role, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from "discord.js";

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName("addstreamer")
    .setDescription("Ajouter un streamer à suivre")
    .addStringOption(option =>
        option
            .setName("pseudo")
            .setDescription("Le pseudo du streamer à suivre")
            .setRequired(true)
    )
    .addRoleOption(option =>
        option
            .setName("role")
            .setDescription("Le rôle à mentionner lorsqu'un streamer commence à streamer (optionnel)")
            .setRequired(false)
    )

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ ephemeral: true, fetchReply: true })
    if (!await hasPermission(interaction, [PermissionFlagsBits.ManageChannels], false)) {
        await interaction.editReply({ embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de changer la configuration."))] })
        return
    }
    const pseudo = interaction.options.get("pseudo")?.value as string
    const role = interaction.options.get("role") as Role | null

    // Add streamer to database
    const databaseCheck = await prisma.streams.findFirst({
        where: {
            AND: [
                { channelName: pseudo },
                { guildId: interaction.guildId as string }
            ]
        }
    })

    if (databaseCheck) {
        await interaction.editReply({ embeds: [errorEmbed(interaction, new Error("Ce streamer est déjà suivi."))] })
        return
    }

    await prisma.streams.create({
        data: {
            channelName: pseudo,
            guildId: interaction.guildId as string,
            roleId: role?.id || null,
            discordChannelId: interaction.channelId as string,
            messageSentId: "", // Provide a valid message ID if available
        }
    })

    streamRetriever.addStream(pseudo)
    await interaction.editReply({ content: `Le streamer ${pseudo} a été ajouté à la liste des streamers à suivre.` })
}