import { CommandInteraction, ContextMenuCommandInteraction, User } from 'discord.js';
import Jimp from 'jimp';
import { prisma } from './database';

const background = 'assets/img/quote.png';
const smoke = 'assets/img/smoke.png';
const fontPath = 'assets/fonts/Ubuntu.fnt';

export async function createQuote(
    interaction: CommandInteraction | ContextMenuCommandInteraction,
    quote: string,
    author: User,
    context?: string,
    date?: string,
    userProfilePicture?: string
): Promise<Buffer> {
    const image = await Jimp.read(background);
    const jimpQuoteFont = await Jimp.loadFont(fontPath);
    const otherThingsFont = await Jimp.loadFont(Jimp.FONT_SANS_32_WHITE);
    const contextFont = await Jimp.loadFont(Jimp.FONT_SANS_16_WHITE);

    if (userProfilePicture) {
        const maxPictureHeight = image.bitmap.height - 18;
        const profilePicture = await Jimp.read(userProfilePicture);
        profilePicture.opacity(0.3);
        image.composite(profilePicture.resize(Jimp.AUTO, maxPictureHeight), 11, 9);
    }

    const smokeImage = await Jimp.read(smoke);
    image.composite(
        smokeImage.resize(image.bitmap.width, Jimp.AUTO),
        0,
        image.bitmap.height - smokeImage.bitmap.height
    );

    const maxQuoteWidth = image.bitmap.width - 550;
    const quoteYPosition = (image.bitmap.height - Jimp.measureTextHeight(jimpQuoteFont, quote, maxQuoteWidth)) / 2;
    image.print(jimpQuoteFont, 350, quoteYPosition, '"' + quote + '"', maxQuoteWidth);
    if (context) {
        const maxContextWidth = image.bitmap.width - 550;
        const contextYPosition =
            quoteYPosition + Jimp.measureTextHeight(jimpQuoteFont, '"' + quote + '"', maxQuoteWidth) + 20;
        image.print(contextFont, 350, contextYPosition, context, maxContextWidth);
    }
    const authorXPosition =
        image.bitmap.width -
        40 -
        Jimp.measureText(otherThingsFont, '@' + author?.displayName + ' - ' + date || 'Anonyme - ' + date);
    image.print(otherThingsFont, authorXPosition, 360, '@' + author?.displayName + ' - ' + date || 'Anonyme - ' + date);

    const buffer = await image.getBufferAsync(Jimp.MIME_PNG);

    await prisma.quote.create({
        data: {
            quote: quote,
            author: {
                connectOrCreate: {
                    where: {
                        userId: author?.id,
                    },
                    create: {
                        userId: author?.id,
                    },
                },
            },
            context: context,
            guildId: interaction.guildId ?? '0',
            createdAt: (interaction.options.get('date')?.value as string) || new Date().toISOString(),
        },
    });

    return buffer;
}
