import { ChatInputCommandInteraction, MessageFlags, SlashCommandBuilder, TextChannel } from 'discord.js';
import type { ICommand } from '../command';
import { generateQuoteImage, insertQuoteInDatabase } from '../../../utils/quoteMaker';
import { prisma } from '../../../utils/core/database';
import { successEmbedGenerator } from '../../utils/embeds';
import { getStringOption, getUserOption } from '../../utils/commandOptions';

export const quote: ICommand = {
    data: new SlashCommandBuilder()
        .setName('quote')
        .setDescription('Créer une citation')
        .addStringOption((option) => option.setName('citation').setDescription('La citation').setRequired(true))
        .addUserOption((option) => option.setName('auteur').setDescription("L'auteur de la citation").setRequired(true))
        .addStringOption((option) =>
            option.setName('contexte').setDescription('[Optionnel] Le contexte de la citation').setRequired(false)
        ),
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            withResponse: true,
            flags: [MessageFlags.Ephemeral],
        });
        const quote = getStringOption(interaction, 'citation', true);
        const author = getUserOption(interaction, 'auteur', true);
        const context = getStringOption(interaction, 'contexte');
        const date = new Date().toLocaleDateString('fr-FR');
        const userProfilePicture = author.displayAvatarURL({ size: 512, extension: 'png', forceStatic: true });

        let channelWhereToPost = null;
        if (interaction.guildId != null) {
            channelWhereToPost = await prisma.config.findFirst({
                where: {
                    guildId: interaction.guildId,
                    key: 'quoteChannel',
                },
            });
        }
        let channel: TextChannel;
        if (channelWhereToPost) {
            channel = interaction.guild?.channels.cache.get(channelWhereToPost.value) as TextChannel;
        } else {
            channel = interaction.channel as TextChannel;
        }
        const imageBuffer = await generateQuoteImage(
            quote,
            author.displayName,
            date,
            userProfilePicture,
            context ?? undefined
        );

        await insertQuoteInDatabase(quote, author.id, context ?? undefined, interaction.guildId ?? undefined);

        const messageSent = await channel.send({
            files: [imageBuffer],
            content: `"${quote}" - ${author?.toString() ?? 'Anonyme'} - ${date} ${context ? `\n*${context}*` : ''}`,
        });
        await interaction.editReply({
            embeds: [successEmbedGenerator(`Citation créée et envoyée ${messageSent.url}`)],
        });
    },
};
