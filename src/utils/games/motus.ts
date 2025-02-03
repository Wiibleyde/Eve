import { ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder } from 'discord.js';
import { client } from '../..';
import { backSpace } from '../textUtils';

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
        attempt = attempt
            .normalize('NFD')
            .replace(/[\u0300-\u036f]/g, '')
            .toUpperCase();
        if (this.state !== GameState.PLAYING) return TryReturn.INVALID;
        if (attempt.length !== this.wordLength) return TryReturn.INVALID;
        if (this.tries.length === MotusGame.maxAttempts) return TryReturn.INVALID;

        const attemptArray = attempt.toUpperCase().split('');
        const attemptState: MotusLetterState[] = new Array(this.wordLength).fill(MotusLetterState.EMPTY);

        const letterCount: { [key: string]: number } = {};
        this.wordArray.forEach((letter) => {
            letterCount[letter] = (letterCount[letter] || 0) + 1;
        });

        for (let i = 0; i < this.wordLength; i++) {
            if (this.wordArray[i] === attemptArray[i]) {
                attemptState[i] = MotusLetterState.FOUND;
                letterCount[attemptArray[i]]--;
            }
        }

        for (let i = 0; i < this.wordLength; i++) {
            if (attemptState[i] === MotusLetterState.EMPTY) {
                if (letterCount[attemptArray[i]] > 0) {
                    attemptState[i] = MotusLetterState.MISPLACED;
                    letterCount[attemptArray[i]]--;
                } else {
                    attemptState[i] = MotusLetterState.NOT_FOUND;
                }
            }
        }

        for (let i = 0; i < this.wordLength; i++) {
            if (attemptState[i] === MotusLetterState.EMPTY) {
                attemptState[i] = MotusLetterState.NOT_FOUND;
            }
        }

        this.tries.push(attemptState);
        this.attempts.push(attempt.toUpperCase());
        this.userAttempts.push(userId);

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
        return this.tries[this.tries.length - 1].every((state) => state === MotusLetterState.FOUND);
    }

    private isLosingAttempt(): boolean {
        return this.tries.length === MotusGame.maxAttempts;
    }

    public async getEmbed(): Promise<{
        embed: EmbedBuilder;
        components: ActionRowBuilder<ButtonBuilder>[];
        attachments: { attachment: string; name: string }[];
    }> {
        let color = 0x00ff00;
        switch (this.state) {
            case GameState.WON:
                color = 0x00ff00;
                break;
            case GameState.LOST:
                color = 0xff0000;
                break;
            case GameState.PLAYING:
                // Make step 1, 2, 3
                if (this.tries.length <= 2) {
                    color = 0x00ff00;
                } else if (this.tries.length <= 4) {
                    color = 0xffff00;
                } else {
                    color = 0xff0000;
                }
                break;
        }

        const embed = new EmbedBuilder()
            .setTitle('Motus')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() || '' })
            .setTimestamp()
            .setThumbnail('attachment://eve-motus.png')
            .setColor(color);

        if (this.state === GameState.WON) {
            embed.setDescription(`Bravo, vous avez trouvé le mot en ${this.tries.length} essais !`);
        } else if (this.state === GameState.LOST) {
            embed.setDescription(
                `Dommage, vous n'avez pas trouvé le mot en ${this.tries.length} essais. Le mot était ${'`'}${this.word}${'`'}.`
            );
        } else {
            embed.setDescription(
                `Lancé par <@${this.originUserId}>, à vous de jouer !${backSpace}Le mot à trouver contient ${this.wordLength} lettres et commence par ${'`'}${this.word[0]}${'`'}.`
            );
        }

        for (let i = 0; i < MotusGame.maxAttempts; i++) {
            const attempt = this.tries[i];
            if (!attempt) {
                break;
            }
            const attemptString = this.attempts[i]
                .split('')
                .map((letter) => ` ${letter} `)
                .join('');
            const emojiString = attempt
                .map((state) => {
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
                })
                .join(' ');

            const username = await client.users.fetch(this.userAttempts[i]);

            embed.addFields({
                name: `Essai ${i + 1} par ${username.displayName}`,
                value: `${'`'}${attemptString}${'`'}${backSpace}${emojiString}`,
            });
        }

        const buttons = [
            new ButtonBuilder().setCustomId('handleMotusTry').setLabel('Essayer').setStyle(ButtonStyle.Primary),
            // new ButtonBuilder().setCustomId("handleMotusGiveUp").setLabel("Abandonner").setStyle(ButtonStyle.Danger)
        ];

        const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

        return {
            embed,
            components: this.state === GameState.PLAYING ? [actionRow] : [],
            attachments: [{ attachment: motusLogo, name: 'eve-motus.png' }],
        };
    }

    public endGame(state: GameState) {
        this.state = state;
    }
}

export const games = new Map<string, MotusGame>(); // Map<messageId, MotusGame>

export async function getRandomWord(): Promise<string> {
    const url = 'https://trouve-mot.fr/api/random';
    const response = await fetch(url);
    const data = await response.json();
    return data[0].name
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .toUpperCase();
}
