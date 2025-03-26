// import { streamRetriever } from "@/index";
import { prisma } from '@/utils/database';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';
import { CommandInteraction, MessageFlags, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('addstreamer')
    .setDescription('Ajouter des streams à la liste des streams suivis')
    .addStringOption((option) => option.setName('stream').setDescription('Nom du streamer à ajouter').setRequired(true))
    .addChannelOption((option) =>
        option.setName('channel').setDescription('Salon où envoyer les notifications').setRequired(true)
    )
    .addRoleOption((option) =>
        option.setName('role').setDescription('Rôle à mentionner lors de la notification').setRequired(false)
    );

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission d'ajouter des streams"))],
        });
        return;
    }
    const streamer = (interaction.options.get('stream')?.value as string).toLowerCase();
    const channel = interaction.options.get('channel')?.value as string;
    const role = interaction.options.get('role')?.value as string | null;

    const isAlreadyFollowing = await prisma.stream.findFirst({
        where: {
            AND: [
                {
                    twitchChannelName: streamer,
                },
                {
                    guildId: interaction.guildId as string,
                },
            ],
        },
    });
    if (isAlreadyFollowing) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error('Ce streamer est déjà suivi sur ce serveur'))],
        });
        return;
    }

    await prisma.stream.create({
        data: {
            twitchChannelName: streamer,
            guildId: interaction.guildId as string,
            channelId: channel,
            roleId: role || null,
        },
    });

    await interaction.editReply({
        embeds: [
            successEmbed(
                interaction,
                `Le streamer ${streamer} a bien été ajouté à la liste des streams suivis  (le message peut mettre un peu de temps à arriver)`
            ),
        ],
    });
}
