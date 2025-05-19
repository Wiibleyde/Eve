import * as fs from 'fs';
import * as path from 'path';

export interface LoggerOptions {
    logDir?: string;
    minLevel?: 'debug' | 'info' | 'warn' | 'error';
    format?: (level: string, message: string) => string;
}

const LEVELS = { debug: 0, info: 1, warn: 2, error: 3 };

// Codes couleur ANSI pour la console
const COLORS = {
    reset: '\x1b[0m',
    debug: '\x1b[32m', // Vert
    info: '\x1b[34m', // Bleu
    warn: '\x1b[33m', // Jaune
    error: '\x1b[31m', // Rouge
};

export class Logger {
    private logDir: string;
    private minLevel: number;
    private formatFn: (level: string, message: string) => string;

    private constructor(options: LoggerOptions = {}) {
        this.logDir = options.logDir || path.join(__dirname, '..', 'logs');
        this.minLevel = LEVELS[options.minLevel || 'debug']; // Changé 'info' en 'debug' ici
        this.formatFn =
            options.format ||
            ((level, message) => {
                const timestamp = new Date().toISOString();
                return `[${timestamp}] [${level.toUpperCase()}] ${message}\n`;
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
        const logMessage = this.formatFn(level, message);
        fs.appendFileSync(this.getLogFilePath(), logMessage, { encoding: 'utf8' });

        // Affiche aussi dans la console avec des couleurs
        const colorCode = COLORS[level];
        const coloredMessage = `${colorCode}${logMessage.trim()}${COLORS.reset}`;

        if (level === 'error') {
            console.error(coloredMessage);
        } else if (level === 'warn') {
            console.warn(coloredMessage);
        } else if (level === 'info') {
            console.info(coloredMessage);
        } else if (level === 'debug') {
            console.debug(coloredMessage);
        }
    }

    debug(...args: any[]) {
        this.writeLog('debug', this.formatArgs(args));
    }

    info(...args: any[]) {
        this.writeLog('info', this.formatArgs(args));
    }

    warn(...args: any[]) {
        this.writeLog('warn', this.formatArgs(args));
    }

    error(...args: any[]) {
        this.writeLog('error', this.formatArgs(args));
    }

    private formatArgs(args: any[]): string {
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
