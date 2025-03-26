import {
    CommandInteraction,
    MessageFlags,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
    TextChannel,
} from 'discord.js';
import { prisma } from '@/utils/database';
import { successEmbed, errorEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { createQuote } from '@/utils/quote';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('quote')
    .setDescription('Créer une citation')
    .addStringOption((option) => option.setName('citation').setDescription('La citation').setRequired(true))
    .addUserOption((option) => option.setName('auteur').setDescription("L'auteur de la citation").setRequired(true))
    .addStringOption((option) =>
        option.setName('contexte').setDescription('[Optionnel] Le contexte de la citation').setRequired(false)
    )
    .addStringOption((option) =>
        option.setName('date').setDescription('[Optionnel] La date de la citation').setRequired(false)
    );

/**
 * Executes the quote command, which processes and posts a quote to a specified channel.
 *
 * @param interaction - The command interaction object containing the user's input.
 * @returns A promise that resolves when the command execution is complete.
 *
 * The function performs the following steps:
 * 1. Defers the reply to the interaction.
 * 2. Retrieves the quote, author, context, and date from the interaction options.
 * 3. Determines the channel where the quote should be posted.
 * 4. Processes the author's profile picture and overlays it on a background image.
 * 5. Creates a new quote record in the database.
 * 6. Composites additional images and text onto the background image.
 * 7. Sends the final image to the determined channel.
 * 8. Edits the original interaction reply with a success message.
 *
 * If the author is not found, the function replies with an error message.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const quote = interaction.options.get('citation')?.value as string;
    const author = interaction.options.get('auteur')?.user;
    const context = interaction.options.get('contexte')?.value as string;
    const date = (interaction.options.get('date')?.value as string) || new Date().toLocaleDateString();
    const userProfilePicture = author?.displayAvatarURL({ extension: 'png', size: 1024 });

    if (!author) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("L'auteur de la citation n'a pas été trouvé"))],
        });
        return;
    }

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

    const buffer = await createQuote(interaction, quote, author, context, date, userProfilePicture);

    const messageSent = await channel.send({
        files: [buffer],
        content: `"${quote}" - ${author?.toString() ?? 'Anonyme'} - ${date} ${context ? `${backSpace}${context}` : ''}`,
    });
    await interaction.editReply({
        embeds: [successEmbed(interaction, `Citation créée et envoyée ${messageSent.url}`)],
    });
}
