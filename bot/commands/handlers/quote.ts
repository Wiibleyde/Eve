import { MessageFlags, SlashCommandBuilder, TextChannel, User } from "discord.js";
import type { ICommand } from "../command";
import { generateQuoteImage, insertQuoteInDatabase } from "../../../utils/quoteMaker";
import { prisma } from "../../../utils/database";
import { successEmbedGenerator } from "../../utils/embeds";
import { logger } from "../../..";

export const quote: ICommand = {
    data: new SlashCommandBuilder()
        .setName("quote")
        .setDescription("Créer une citation")
        .addStringOption((option) => option.setName('citation').setDescription('La citation').setRequired(true))
        .addUserOption((option) => option.setName('auteur').setDescription("L'auteur de la citation").setRequired(true))
        .addStringOption((option) =>
            option.setName('contexte').setDescription('[Optionnel] Le contexte de la citation').setRequired(false)
        ),
    execute: async (interaction) => {
        await interaction.deferReply({
            withResponse: true,
            flags: [MessageFlags.Ephemeral],
        });
        const quote = interaction.options.get('citation')?.value as string;
        const author = interaction.options.get('auteur')?.user as User;
        const context = interaction.options.get('contexte')?.value as string;
        const date = new Date().toLocaleDateString('fr-FR');
        const userProfilePicture = author.displayAvatarURL({ size: 512, extension: 'png', forceStatic: true });

        logger.debug('quote:', {
            quote,
            author: author.displayName,
            context,
            date,
            userProfilePicture,
        });

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
            context
        );

        await insertQuoteInDatabase(quote, author.id, context, interaction.guildId ?? undefined)


        const messageSent = await channel.send({
            files: [imageBuffer],
            content: `"${quote}" - ${author?.toString() ?? 'Anonyme'} - ${date} ${context ? `\n*${context}*` : ''}`,
        });
        await interaction.editReply({
            embeds: [successEmbedGenerator(`Citation créée et envoyée ${messageSent.url}`)],
        });
    }
}