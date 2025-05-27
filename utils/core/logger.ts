import * as fs from 'fs';
import * as path from 'path';
import { WebhookClient, EmbedBuilder } from 'discord.js';
import { client } from '../../bot/bot';

export interface LoggerOptions {
    logDir?: string;
    minLevel?: 'debug' | 'info' | 'warn' | 'error';
    format?: (level: string, message: string, fileInfo?: string) => string;
    showFileInfo?: boolean;
    discordWebhook?: string; // URL complète du webhook Discord
    discordMinLevel?: 'debug' | 'info' | 'warn' | 'error';
}

const LEVELS = { debug: 0, info: 1, warn: 2, error: 3 };

// Codes couleur ANSI pour la console
const COLORS = {
    reset: '\x1b[0m',
    debug: '\x1b[32m', // Vert
    info: '\x1b[34m', // Bleu
    warn: '\x1b[33m', // Jaune
    error: '\x1b[31m', // Rouge
    fileInfo: '\x1b[36m', // Cyan pour les infos de fichier
};

// Couleurs pour Discord (codes couleur hexadécimaux)
const DISCORD_COLORS = {
    debug: 0x00ff00, // Vert
    info: 0x13b8eb, // Bleu
    warn: 0xffff00, // Jaune
    error: 0xff0000, // Rouge
};

export class Logger {
    private logDir: string;
    private minLevel: number;
    private formatFn: (level: string, message: string, fileInfo?: string) => string;
    private showFileInfo: boolean;
    private discordClient?: WebhookClient;
    private discordMinLevel: number;

    private constructor(options: LoggerOptions = {}) {
        this.logDir = options.logDir || path.join(__dirname, '..', 'logs');
        this.minLevel = LEVELS[options.minLevel || 'debug'];
        this.showFileInfo = options.showFileInfo !== undefined ? options.showFileInfo : true;
        this.discordMinLevel = LEVELS[options.discordMinLevel || 'warn']; // Par défaut, uniquement les warn et error

        // Initialiser le client Discord avec l'URL du webhook si fournie
        if (options.discordWebhook) {
            try {
                this.discordClient = new WebhookClient({ url: options.discordWebhook });
            } catch (error) {
                console.error('Erreur lors de la configuration du webhook Discord:', error);
            }
        }

        this.formatFn =
            options.format ||
            ((level, message, fileInfo) => {
                const timestamp = new Date().toISOString();
                const fileInfoStr = fileInfo && this.showFileInfo ? ` (${fileInfo})` : '';
                return `[${timestamp}] [${level.toUpperCase()}]${fileInfoStr} ${message}\n`;
            });

        if (!fs.existsSync(this.logDir)) {
            fs.mkdirSync(this.logDir, { recursive: true });
        }
    }

    private getLogFilePath() {
        const date = new Date();
        const yyyy = date.getFullYear();
        const mm = String(date.getMonth() + 1).padStart(2, '0');
        const dd = String(date.getDate()).padStart(2, '0');
        return path.join(this.logDir, `${yyyy}-${mm}-${dd}.log`);
    }

    private writeLog(level: keyof typeof LEVELS, message: string) {
        if (LEVELS[level] < this.minLevel) return;

        const fileInfo = this.getCallerInfo();
        const logMessage = this.formatFn(level, message, fileInfo);
        fs.appendFileSync(this.getLogFilePath(), logMessage, { encoding: 'utf8' });

        const colorCode = COLORS[level];
        let coloredMessage = `${colorCode}${logMessage.trim()}${COLORS.reset}`;

        if (fileInfo && this.showFileInfo) {
            coloredMessage = coloredMessage.replace(`(${fileInfo})`, `${COLORS.fileInfo}(${fileInfo})${colorCode}`);
        }

        if (level === 'error') {
            console.error(coloredMessage);
        } else if (level === 'warn') {
            console.warn(coloredMessage);
        } else if (level === 'info') {
            console.info(coloredMessage);
        } else if (level === 'debug') {
            console.debug(coloredMessage);
        }

        if (this.discordClient && LEVELS[level] >= this.discordMinLevel) {
            this.sendToDiscord(level, message, fileInfo);
        }
    }

    private async sendToDiscord(level: keyof typeof LEVELS, message: string, fileInfo?: string): Promise<void> {
        if (!this.discordClient) return;

        try {
            const timestamp = new Date();
            const fileInfoStr = fileInfo && this.showFileInfo ? `\n**Source:** ${fileInfo}` : '';

            const iconUrl = client.user?.displayAvatarURL();

            const embed = new EmbedBuilder()
                .setTitle(`${level.toUpperCase()}`)
                .setDescription(message + fileInfoStr)
                .setColor(DISCORD_COLORS[level])
                .setTimestamp(timestamp)
                .setFooter({ text: `Eve - Logs`, iconURL: iconUrl });

            this.discordClient
                .send({
                    username: `Eve - ${level.toUpperCase()}`,
                    avatarURL: iconUrl,
                    embeds: [embed],
                })
                .catch((error) => {
                    console.error("Erreur lors de l'envoi du message Discord:", error);
                });
        } catch (error) {
            console.error('Erreur lors de la préparation du message Discord:', error);
        }
    }

    private getCallerInfo(): string | undefined {
        if (!this.showFileInfo) return undefined;

        try {
            const stackLines = new Error().stack?.split('\n') || [];
            // Skip first 3 lines (Error, getCallerInfo, writeLog, and the actual logger method)
            const callerLine = stackLines[4];

            if (callerLine) {
                // Formats like: at functionName (/path/to/file.js:line:column)
                const match = callerLine.match(/\((.+?):(\d+):\d+\)/) || callerLine.match(/at (.+?):(\d+):\d+/);

                if (match) {
                    const filePath = match[1];
                    const lineNumber = match[2];
                    // Get just the filename from the path
                    if (filePath) {
                        const fileName = path.basename(filePath);
                        return `${fileName}:${lineNumber}`;
                    }
                }
            }
        } catch {
            return undefined;
        }

        return undefined;
    }

    debug(...args: unknown[]) {
        this.writeLog('debug', this.formatArgs(args));
    }

    info(...args: unknown[]) {
        this.writeLog('info', this.formatArgs(args));
    }

    warn(...args: unknown[]) {
        this.writeLog('warn', this.formatArgs(args));
    }

    error(...args: unknown[]) {
        this.writeLog('error', this.formatArgs(args));
    }

    private formatArgs(args: unknown[]): string {
        // Imiter le comportement de console.log pour formater plusieurs arguments
        return args.map((arg) => (typeof arg === 'object' ? JSON.stringify(arg) : String(arg))).join(' ');
    }

    static init(options?: LoggerOptions | string) {
        if (typeof options === 'string') {
            return new Logger({ logDir: options });
        }
        return new Logger(options);
    }
}
