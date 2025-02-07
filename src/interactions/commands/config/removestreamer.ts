import { streamRetriever } from "@/index";
import { prisma } from "@/utils/database";
import { errorEmbed, successEmbed } from "@/utils/embeds";
import { hasPermission } from "@/utils/permissionTester";
import { CommandInteraction, MessageFlags, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from "discord.js";

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName("removestreamer")
    .setDescription("Supprimer un stream de la liste des streams suivis")
    .addStringOption((option) => option.setName("stream").setDescription("Nom du streamer à supprimer").setRequired(true))

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission d'ajouter des streams"))],
        });
        return
    }
    const streamer = (interaction.options.get("stream")?.value as string).toLowerCase();

    const isFollowing = await prisma.stream.findFirst({
        where: {
            AND: [
                {
                    twitchChannelName: streamer,
                },
                {
                    guildId: interaction.guildId as string,
                },
            ],
        }
    });
    if (!isFollowing) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Ce streamer n'est pas suivi sur ce serveur"))],
        });
        return;
    }
    await streamRetriever.removeStream(streamer, interaction.guildId as string);

    await interaction.editReply({
        embeds: [successEmbed(interaction, `Le streamer ${streamer} a bien été retiré de la liste des streams suivis`)],
    });
}