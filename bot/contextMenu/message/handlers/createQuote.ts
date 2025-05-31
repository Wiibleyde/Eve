import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    MessageFlags,
    MessageContextMenuCommandInteraction,
    TextChannel,
} from 'discord.js';
import { prisma } from '../../../../utils/core/database';
import { generateQuoteImage, insertQuoteInDatabase } from '../../../../utils/quoteMaker';
import { successEmbedGenerator } from '../../../utils/embeds';
import type { IContextMenuMessageCommand } from '../contextMenuMessage';

export const createQuote: IContextMenuMessageCommand = {
    data: new ContextMenuCommandBuilder().setName('Créer une citation').setType(ApplicationCommandType.Message),
    async execute(interaction: MessageContextMenuCommandInteraction) {
        await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
        const quote = interaction.targetMessage?.content;
        const author = interaction.targetMessage?.author;
        const date = interaction.targetMessage?.createdAt.toLocaleDateString('fr-FR');
        const userProfilePicture = author?.displayAvatarURL({ size: 512, extension: 'png', forceStatic: true });

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
        const imageBuffer = await generateQuoteImage(quote, author.displayName, date, userProfilePicture);

        await insertQuoteInDatabase(quote, author.id, undefined, interaction.guildId ?? undefined);

        const messageSent = await channel.send({
            files: [imageBuffer],
            content: `"${quote}" - ${author?.toString() ?? 'Anonyme'} - ${date}`,
        });
        await interaction.editReply({
            embeds: [successEmbedGenerator(`Citation créée et envoyée ${messageSent.url}`)],
        });
    },
};
