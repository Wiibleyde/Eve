import { Jimp, JimpMime, loadFont, measureText, measureTextHeight, ResizeStrategy } from 'jimp';
import { prisma } from './core/database';

const background = 'assets/img/quote.png';
const smoke = 'assets/img/smoke.png';

const fontPath = 'assets/fonts/ubuntu/Ubuntu.fnt';
const sans32WhitePath = 'assets/fonts/open-sans/open-sans-32-white/open-sans-32-white.fnt';
const sans16WhitePath = 'assets/fonts/open-sans/open-sans-16-white/open-sans-16-white.fnt';

export async function generateQuoteImage(
    quote: string,
    author: string,
    date: string,
    userProfilePicture: string,
    context?: string
): Promise<Buffer> {
    if (!quote) {
        throw new Error('Quote cannot be empty');
    }
    if (quote.length > 200) {
        throw new Error('Quote cannot exceed 200 characters');
    }
    if (context && context.length > 200) {
        throw new Error('Context cannot exceed 200 characters');
    }
    const image = await Jimp.read(background);
    const jimpQuoteFont = await loadFont(fontPath);
    const otherThingsFont = await loadFont(sans32WhitePath);
    const contextFont = await loadFont(sans16WhitePath);

    if (userProfilePicture) {
        const maxPictureHeight = image.bitmap.height - 18;
        const profilePicture = await Jimp.read(userProfilePicture);
        profilePicture.opacity(0.3);
        image.composite(profilePicture.resize({ h: maxPictureHeight, mode: ResizeStrategy.BEZIER }), 11, 9);
    }

    const smokeImage = await Jimp.read(smoke);
    image.composite(smokeImage.resize({ w: image.bitmap.width }), 0, image.bitmap.height - smokeImage.bitmap.height);

    const maxQuoteWidth = image.bitmap.width - 550;
    const quoteYPosition = (image.bitmap.height - measureTextHeight(jimpQuoteFont, quote, maxQuoteWidth)) / 2;
    image.print({
        x: 350,
        y: quoteYPosition,
        font: jimpQuoteFont,
        text: '"' + quote + '"',
        maxWidth: maxQuoteWidth,
    });
    if (context) {
        const maxContextWidth = image.bitmap.width - 550;
        const contextYPosition =
            quoteYPosition + measureTextHeight(jimpQuoteFont, '"' + quote + '"', maxQuoteWidth) + 20;
        image.print({
            x: 350,
            y: contextYPosition,
            font: contextFont,
            text: context,
            maxWidth: maxContextWidth,
        });
    }
    const authorXPosition =
        image.bitmap.width - 40 - measureText(otherThingsFont, '@' + author + ' - ' + date || 'Anonyme - ' + date);
    image.print({
        x: authorXPosition,
        y: 360,
        font: otherThingsFont,
        text: '@' + author + ' - ' + date || 'Anonyme - ' + date,
    });

    const buffer = await image.getBuffer(JimpMime.png);

    return buffer;
}

export async function insertQuoteInDatabase(
    quote: string,
    authorId: string,
    context?: string,
    guildId?: string
): Promise<void> {
    // Assuming you have a Prisma client instance available as `prisma`
    await prisma.quote.create({
        data: {
            quote: quote,
            author: {
                connectOrCreate: {
                    where: {
                        userId: authorId,
                    },
                    create: {
                        userId: authorId,
                    },
                },
            },
            context: context,
            guildId: guildId ?? '0',
            createdAt: new Date(),
        },
    });
}
