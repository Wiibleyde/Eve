import { CommandInteraction, ContextMenuCommandInteraction, User } from 'discord.js';
import { Jimp, JimpMime, loadFont, measureText, measureTextHeight, ResizeStrategy } from 'jimp';
import { prisma } from './database';

const background = 'assets/img/quote.png';
const smoke = 'assets/img/smoke.png';

const fontPath = 'assets/fonts/ubuntu/Ubuntu.fnt';
const sans32WhitePath = 'assets/fonts/open-sans/open-sans-32-white/open-sans-32-white.fnt';
const sans16WhitePath = 'assets/fonts/open-sans/open-sans-16-white/open-sans-16-white.fnt';

export async function createQuote(
    interaction: CommandInteraction | ContextMenuCommandInteraction,
    quote: string,
    author: User,
    context?: string,
    date?: string,
    userProfilePicture?: string
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
    quote = await fixMessage(interaction, quote);
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
        image.bitmap.width -
        40 -
        measureText(otherThingsFont, '@' + author?.displayName + ' - ' + date || 'Anonyme - ' + date);
    image.print({
        x: authorXPosition,
        y: 360,
        font: otherThingsFont,
        text: '@' + author?.displayName + ' - ' + date || 'Anonyme - ' + date,
    });

    const buffer = await image.getBuffer(JimpMime.png);

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

async function fixMessage(
    interaction: CommandInteraction | ContextMenuCommandInteraction,
    message: string
): Promise<string> {
    // Remove the "" if they are at the beginning and end of the message
    message = message.replace(/^"(.+)"$/, '$1');
    // Replace <@123456789012345678> with @username
    message = message.replace(/<@!?(\d+)>/g, (match, userId) => {
        const user = interaction.client.users.cache.get(userId);
        return user ? `@${user.username}` : match;
    });
    // Replace <#123456789012345678> with #channel
    message = message.replace(/<#(\d+)>/g, (match, channelId) => {
        const channel = interaction.guild?.channels.cache.get(channelId);
        return channel ? `#${channel.name}` : match;
    });
    // Replace <@&123456789012345678> with @role
    message = message.replace(/<@&(\d+)>/g, (match, roleId) => {
        const role = interaction.guild?.roles.cache.get(roleId);
        return role ? `@${role.name}` : match;
    });
    // Replace <a:emoji_name:123456789012345678> with :emoji_name:
    message = message.replace(/<a?:(\w+):(\d+)>/g, (match, emojiName) => {
        return `:${emojiName}:`;
    });
    // Replace <t:1234567890> with the date
    message = message.replace(/<t:(\d+)>/g, (match, timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        return date.toLocaleDateString('fr-FR', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
        });
    });
    // Replace <t:1234567890:T> with the date and time
    message = message.replace(/<t:(\d+):T>/g, (match, timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        return date.toLocaleString('fr-FR', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
        });
    });
    // Replace <t:1234567890:F> with the date and time in full format
    message = message.replace(/<t:(\d+):F>/g, (match, timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        return date.toLocaleString('fr-FR', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
        });
    });
    // Replace <t:1234567890:R> with the time ago
    message = message.replace(/<t:(\d+):R>/g, (match, timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        const now = new Date();
        const diff = Math.abs(now.getTime() - date.getTime());
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);
        if (seconds < 60) return `${seconds} secondes`;
        if (minutes < 60) return `${minutes} minutes`;
        if (hours < 24) return `${hours} heures`;
        return `${days} jours`;
    });
    // Replace <t:1234567890:D> with the date in short format
    message = message.replace(/<t:(\d+):D>/g, (match, timestamp) => {
        const date = new Date(parseInt(timestamp) * 1000);
        return date.toLocaleDateString('fr-FR', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
        });
    });
    return message;
}
