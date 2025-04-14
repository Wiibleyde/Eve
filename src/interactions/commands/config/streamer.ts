import { prisma } from '@/utils/database';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { hasPermission } from '@/utils/permissionTester';
import { removeStream } from '@/utils/streams';
import {
    ChatInputCommandInteraction,
    EmbedBuilder,
    MessageFlags,
    SlashCommandBuilder,
    SlashCommandSubcommandBuilder,
} from 'discord.js';

export const data = new SlashCommandBuilder()
    .setName('streamer')
    .setDescription('Gérer les streams suivis par Eve')
    .addSubcommand((subcommand: SlashCommandSubcommandBuilder) =>
        subcommand
            .setName('add')
            .setDescription('Ajouter un streamer à la liste des streams suivis')
            .addStringOption((option) =>
                option.setName('stream').setDescription('Nom du streamer à ajouter').setRequired(true)
            )
            .addChannelOption((option) =>
                option.setName('channel').setDescription('Salon où envoyer les notifications').setRequired(true)
            )
            .addRoleOption((option) =>
                option.setName('role').setDescription('Rôle à mentionner lors de la notification').setRequired(false)
            )
    )
    .addSubcommand((subcommand: SlashCommandSubcommandBuilder) =>
        subcommand
            .setName('remove')
            .setDescription('Supprimer un streamer de la liste des streams suivis')
            .addStringOption((option) =>
                option.setName('stream').setDescription('Nom du streamer à supprimer').setRequired(true)
            )
    )
    .addSubcommand((subcommand: SlashCommandSubcommandBuilder) =>
        subcommand.setName('list').setDescription('Afficher la liste des streamers suivis sur ce serveur')
    );

export async function execute(interaction: ChatInputCommandInteraction): Promise<void> {
    await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });

    if (!(await hasPermission(interaction, [], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de gérer les streams"))],
        });
        return;
    }

    const subcommand = interaction.options.getSubcommand();

    if (subcommand === 'add') {
        await handleAddStreamer(interaction);
    } else if (subcommand === 'remove') {
        await handleRemoveStreamer(interaction);
    } else if (subcommand === 'list') {
        await handleListStreamers(interaction);
    }
}

async function handleAddStreamer(interaction: ChatInputCommandInteraction): Promise<void> {
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
                `Le streamer ${streamer} a bien été ajouté à la liste des streams suivis (le message peut mettre un peu de temps à arriver)`
            ),
        ],
    });
}

async function handleRemoveStreamer(interaction: ChatInputCommandInteraction): Promise<void> {
    const streamer = (interaction.options.get('stream')?.value as string).toLowerCase();

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
        },
    });
    if (!isFollowing) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Ce streamer n'est pas suivi sur ce serveur"))],
        });
        return;
    }

    await removeStream(streamer, interaction.guildId as string);

    await interaction.editReply({
        embeds: [successEmbed(interaction, `Le streamer ${streamer} a bien été retiré de la liste des streams suivis`)],
    });
}

async function handleListStreamers(interaction: ChatInputCommandInteraction): Promise<void> {
    const streamers = await prisma.stream.findMany({
        where: {
            guildId: interaction.guildId as string,
        },
        select: {
            twitchChannelName: true,
            channelId: true,
            roleId: true,
        },
    });

    if (streamers.length === 0) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Aucun streamer n'est suivi sur ce serveur"))],
        });
        return;
    }

    const embed = new EmbedBuilder()
        .setTitle('Liste des streamers suivis')
        .setDescription(`${streamers.length} streamer(s) suivi(s) sur ce serveur.`)
        .setColor('Blue')
        .setTimestamp();

    for (const streamer of streamers) {
        let roleInfo = 'Aucun rôle mentionné';
        if (streamer.roleId) {
            roleInfo = `<@&${streamer.roleId}>`;
        }

        embed.addFields({
            name: streamer.twitchChannelName,
            value: `Canal: <#${streamer.channelId}>\nRôle: ${roleInfo}`,
            inline: true,
        });
    }

    await interaction.editReply({ embeds: [embed] });
}
