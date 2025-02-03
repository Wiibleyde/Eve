import { prisma } from '@/utils/database';
import { successEmbed } from '@/utils/embeds';
import { createQuote } from '@/utils/quote';
import {
    ApplicationCommandType,
    ContextMenuCommandBuilder,
    MessageContextMenuCommandInteraction,
    TextChannel,
} from 'discord.js';

export const data: ContextMenuCommandBuilder = new ContextMenuCommandBuilder()
    .setName('Créer un citation')
    //@ts-expect-error - This is a valid type
    .setType(ApplicationCommandType.Message);

export async function execute(interaction: MessageContextMenuCommandInteraction): Promise<void> {
    await interaction.deferReply({ ephemeral: true });

    const quote = interaction.targetMessage?.content;
    const author = interaction.targetMessage?.author;
    const date = interaction.targetMessage?.createdAt.toLocaleDateString() || new Date().toLocaleDateString();
    const userProfilePicture = author?.displayAvatarURL({ extension: 'png', size: 1024 });

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

    const buffer = await createQuote(interaction, quote, author, undefined, date, userProfilePicture);

    const messageSent = await channel.send({
        files: [buffer],
        content: `"${quote}" - ${author?.toString() ?? 'Anonyme'} - ${date}`,
    });
    await interaction.editReply({
        embeds: [successEmbed(interaction, `Citation créée et envoyée ${messageSent.url}`)],
    });
}
