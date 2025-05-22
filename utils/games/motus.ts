import { ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder } from 'discord.js';
import { client } from '../../bot/bot';

enum MotusLetterState {
    FOUND,
    MISPLACED,
    NOT_FOUND,
    EMPTY,
}

export enum TryReturn {
    WIN,
    LOSE,
    CONTINUE,
    INVALID,
}

export enum GameState {
    PLAYING,
    WON,
    LOST,
}

const motusLogo = './assets/img/eve-motus.png';

export class MotusGame {
    public state: GameState;
    public readonly wordLength: number;

    private readonly word: string;
    private readonly originUserId: string;
    private readonly wordArray: string[];
    private readonly tries: MotusLetterState[][];
    private readonly attempts: string[];
    private readonly userAttempts: string[];
    private static readonly maxAttempts = 6;

    constructor(word: string, originUserId: string) {
        this.state = GameState.PLAYING;
        this.word = word.toUpperCase();
        this.originUserId = originUserId;
        this.wordArray = this.word.split('');
        this.wordLength = this.word.length;
        this.tries = [];
        this.attempts = [];
        this.userAttempts = [];
    }

    /**
     * Attempts to match the given attempt string with the target word.
     *
     * @param attempt - The string that the user is attempting to match with the target word.
     * @returns Whether the attempt is a winning attempt.
     */
    public tryAttempt(attempt: string, userId: string): TryReturn {
        attempt = this.normalizeString(attempt);

        // Validate attempt
        if (!this.isValidAttempt(attempt)) return TryReturn.INVALID;

        const attemptArray = attempt.split('');
        const attemptState = this.evaluateAttempt(attemptArray);

        // Record the attempt
        this.recordAttempt(attempt, attemptState, userId);

        // Check game state
        return this.updateGameState();
    }

    /**
     * Normalizes a string by removing diacritics and converting to uppercase
     */
    private normalizeString(str: string): string {
        return str
            .normalize('NFD')
            .replace(/[\u0300-\u036f]/g, '')
            .toUpperCase();
    }

    /**
     * Check if the attempt is valid for the current game state
     */
    private isValidAttempt(attempt: string): boolean {
        return (
            this.state === GameState.PLAYING &&
            attempt.length === this.wordLength &&
            this.tries.length < MotusGame.maxAttempts
        );
    }

    /**
     * Evaluates an attempt array against the target word
     */
    private evaluateAttempt(attemptArray: string[]): MotusLetterState[] {
        const attemptState: MotusLetterState[] = new Array(this.wordLength).fill(MotusLetterState.EMPTY);
        const letterCount = this.getLetterCount();

        // First pass: Mark exact matches (FOUND)
        this.markExactMatches(attemptArray, attemptState, letterCount);

        // Second pass: Mark misplaced or not found letters
        this.markRemainingLetters(attemptArray, attemptState, letterCount);

        return attemptState;
    }

    /**
     * Get count of each letter in the target word
     */
    private getLetterCount(): { [key: string]: number } {
        const letterCount: { [key: string]: number } = {};
        this.wordArray.forEach((letter) => {
            letterCount[letter] = (letterCount[letter] || 0) + 1;
        });
        return letterCount;
    }

    /**
     * Mark exact matches in the attempt
     */
    private markExactMatches(
        attemptArray: string[],
        attemptState: MotusLetterState[],
        letterCount: { [key: string]: number }
    ): void {
        for (let i = 0; i < this.wordLength; i++) {
            if (this.wordArray[i] === attemptArray[i]) {
                attemptState[i] = MotusLetterState.FOUND;

                const letter = attemptArray[i];
                if (letter && letterCount[letter] !== undefined) {
                    letterCount[letter]--;
                }
            }
        }
    }

    /**
     * Mark misplaced or not found letters
     */
    private markRemainingLetters(
        attemptArray: string[],
        attemptState: MotusLetterState[],
        letterCount: { [key: string]: number }
    ): void {
        for (let i = 0; i < this.wordLength; i++) {
            if (attemptState[i] === MotusLetterState.EMPTY) {
                const letter = attemptArray[i];

                if (letter && letterCount[letter] && letterCount[letter] > 0) {
                    attemptState[i] = MotusLetterState.MISPLACED;
                    letterCount[letter]--;
                } else {
                    attemptState[i] = MotusLetterState.NOT_FOUND;
                }
            }
        }
    }

    /**
     * Record an attempt in the game history
     */
    private recordAttempt(attempt: string, attemptState: MotusLetterState[], userId: string): void {
        this.tries.push(attemptState);
        this.attempts.push(attempt);
        this.userAttempts.push(userId);
    }

    /**
     * Update game state based on the latest attempt
     */
    private updateGameState(): TryReturn {
        if (this.isWinningAttempt()) {
            this.state = GameState.WON;
            return TryReturn.WIN;
        }

        if (this.isLosingAttempt()) {
            this.state = GameState.LOST;
            return TryReturn.LOSE;
        }

        return TryReturn.CONTINUE;
    }

    private isWinningAttempt(): boolean {
        const lastTry = this.tries[this.tries.length - 1];
        return Array.isArray(lastTry) && lastTry.every((state) => state === MotusLetterState.FOUND);
    }

    private isLosingAttempt(): boolean {
        return this.tries.length === MotusGame.maxAttempts;
    }

    public async getEmbed(): Promise<{
        embed: EmbedBuilder;
        components: ActionRowBuilder<ButtonBuilder>[];
    }> {
        const embed = new EmbedBuilder()
            .setTitle('Motus')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() || '' })
            .setTimestamp()
            .setColor(this.getColorForState());

        embed.setDescription(this.getDescriptionForState());

        // Add attempt fields
        for (let i = 0; i < this.tries.length; i++) {
            const attemptString = this.formatAttemptString(this.attempts[i] ?? '');
            const emojiString = this.tries[i]?.map((state) => this.getEmojiForState(state)).join(' ') ?? '';

            const userId = this.userAttempts[i];
            const displayName = await this.getUserDisplayName(userId ?? '');

            embed.addFields({
                name: `Essai ${i + 1} par ${displayName}`,
                value: `${'`'}${attemptString}${'`'}\n${emojiString}`,
            });
        }

        return {
            embed,
            components: this.getComponents(),
        };
    }

    /**
     * Get the appropriate embed color based on game state and progress
     */
    private getColorForState(): number {
        switch (this.state) {
            case GameState.WON:
                return 0x00ff00;
            case GameState.LOST:
                return 0xff0000;
            case GameState.PLAYING:
                // Color changes as player gets closer to max attempts
                if (this.tries.length <= 2) return 0x00ff00;
                if (this.tries.length <= 4) return 0xffff00;
                return 0xff0000;
        }
    }

    /**
     * Get the description based on game state
     */
    private getDescriptionForState(): string {
        switch (this.state) {
            case GameState.WON:
                return `Bravo, vous avez trouvé le mot en ${this.tries.length} essais !`;
            case GameState.LOST:
                return `Dommage, vous n'avez pas trouvé le mot en ${this.tries.length} essais. Le mot était ${'`'}${this.word}${'`'}.`;
            default:
                return `Lancé par <@${this.originUserId}>, à vous de jouer !\nLe mot à trouver contient ${this.wordLength} lettres et commence par ${'`'}${this.word[0]}${'`'}.`;
        }
    }

    /**
     * Format attempt string with spaces for better visibility
     */
    private formatAttemptString(attempt: string): string {
        return attempt
            .split('')
            .map((letter) => ` ${letter} `)
            .join('');
    }

    /**
     * Get action components based on game state
     */
    private getComponents(): ActionRowBuilder<ButtonBuilder>[] {
        if (this.state !== GameState.PLAYING) return [];

        const buttons = [
            new ButtonBuilder().setCustomId('handleMotusTry').setLabel('Essayer').setStyle(ButtonStyle.Primary),
        ];

        return [new ActionRowBuilder<ButtonBuilder>().addComponents(buttons)];
    }

    public endGame(state: GameState) {
        this.state = state;
    }

    /**
     * Gets the display name of a user by ID
     * @param userId The Discord user ID
     * @returns The user's display name or a fallback
     */
    private async getUserDisplayName(userId: string): Promise<string> {
        if (!userId) return 'Utilisateur';

        try {
            const user = await client.users.fetch(userId);
            return user.displayName || user.username || 'Utilisateur';
        } catch (error) {
            console.error(`Error fetching user ${userId}:`, error);
            return 'Utilisateur';
        }
    }

    /**
     * Gets the emoji representation of a letter state
     * @param state The letter state
     * @returns The corresponding emoji
     */
    private getEmojiForState(state: MotusLetterState): string {
        switch (state) {
            case MotusLetterState.FOUND:
                return '🟩';
            case MotusLetterState.MISPLACED:
                return '🟨';
            case MotusLetterState.NOT_FOUND:
                return '🟥';
            case MotusLetterState.EMPTY:
                return '⬜';
        }
    }
}

export const motusGames = new Map<string, MotusGame>(); // Map<messageId, MotusGame>

export async function getRandomWord(): Promise<string> {
    const url = 'https://trouve-mot.fr/api/random';
    const response = await fetch(url);
    const data = (await response.json()) as { name: string; categorie: string }[];
    if (!data[0] || !data[0].name) {
        throw new Error('No word found in API response');
    }
    return data[0].name
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .toUpperCase();
}
