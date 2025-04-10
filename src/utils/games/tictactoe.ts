import { client } from '@/index';
import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    EmbedBuilder,
    InteractionUpdateOptions,
    MessageCreateOptions,
    MessagePayload,
} from 'discord.js';
import { Jimp, JimpMime, loadFont, measureText, measureTextHeight } from 'jimp';

const fontPath = 'assets/fonts/ubuntu/Ubuntu.fnt';

export enum TicTacToeGameState {
    IN_PROGRESS,
    DRAW,
    X_WINS,
    O_WINS,
}
enum CurrentPlayerState {
    X,
    O,
}

enum CellState {
    EMPTY,
    X,
    O,
}

export class TicTacToeGame {
    private board: CellState[][];
    private winner: string | null;
    private xPlayer: string;
    private oPlayer: string;
    private currentPlayer: CurrentPlayerState;
    private gameState: TicTacToeGameState;

    constructor(xPlayer: string, oPlayer: string) {
        this.board = Array.from({ length: 3 }, () => Array(3).fill(CellState.EMPTY));
        this.winner = null;
        this.xPlayer = xPlayer;
        this.oPlayer = oPlayer;
        this.currentPlayer = CurrentPlayerState.X;
        this.gameState = TicTacToeGameState.IN_PROGRESS;
    }

    public getBoard(): CellState[][] {
        return this.board;
    }

    public getCurrentPlayer(): string {
        return this.currentPlayer === CurrentPlayerState.X ? this.xPlayer : this.oPlayer;
    }

    public getGameState(): TicTacToeGameState {
        return this.gameState;
    }

    public getWinner(): string | null {
        return this.winner;
    }

    public playMove(row: number, col: number, playerId: string): boolean {
        if (this.gameState !== TicTacToeGameState.IN_PROGRESS || this.board[row][col] !== CellState.EMPTY) {
            return false;
        }

        if (
            (this.currentPlayer === CurrentPlayerState.X && playerId !== this.xPlayer) ||
            (this.currentPlayer === CurrentPlayerState.O && playerId !== this.oPlayer)
        ) {
            return false;
        }

        this.board[row][col] = this.currentPlayer === CurrentPlayerState.X ? CellState.X : CellState.O;

        if (this.checkWin()) {
            this.gameState =
                this.currentPlayer === CurrentPlayerState.X ? TicTacToeGameState.X_WINS : TicTacToeGameState.O_WINS;
            this.winner = this.getCurrentPlayer();
        } else if (this.checkDraw()) {
            this.gameState = TicTacToeGameState.DRAW;
        } else {
            this.currentPlayer =
                this.currentPlayer === CurrentPlayerState.X ? CurrentPlayerState.O : CurrentPlayerState.X;
        }
        return true;
    }

    private checkWin(): boolean {
        const playerSymbol = this.currentPlayer === CurrentPlayerState.X ? CellState.X : CellState.O;

        // Check rows
        for (let row = 0; row < 3; row++) {
            if (this.board[row].every((cell) => cell === playerSymbol)) {
                return true;
            }
        }

        // Check columns
        for (let col = 0; col < 3; col++) {
            if (this.board.every((row) => row[col] === playerSymbol)) {
                return true;
            }
        }

        // Check diagonals
        if (
            (this.board[0][0] === playerSymbol &&
                this.board[1][1] === playerSymbol &&
                this.board[2][2] === playerSymbol) ||
            (this.board[0][2] === playerSymbol &&
                this.board[1][1] === playerSymbol &&
                this.board[2][0] === playerSymbol)
        ) {
            return true;
        }

        return false;
    }

    private checkDraw(): boolean {
        return this.board.every((row) => row.every((cell) => cell !== CellState.EMPTY));
    }

    public async getResponse(): Promise<string | MessagePayload | MessageCreateOptions | InteractionUpdateOptions> {
        const imageBuffer = await this.drawImage();
        const embed = new EmbedBuilder();
        if (this.gameState === TicTacToeGameState.X_WINS) {
            embed.setColor(0x00ff00);
            embed.setDescription(`Bravo <@${this.xPlayer}> ! Tu as gagné !`);
        } else if (this.gameState === TicTacToeGameState.O_WINS) {
            embed.setColor(0x00ff00);
            embed.setDescription(`Bravo <@${this.oPlayer}> ! Tu as gagné !`);
        } else if (this.gameState === TicTacToeGameState.DRAW) {
            embed.setColor(0xff0000);
            embed.setDescription(`Match nul !`);
        } else {
            embed.setColor(0x00ff00);
            embed.setDescription(`C'est à <@${this.getCurrentPlayer()}> de jouer !`);
        }

        embed.setTitle('Tic Tac Toe');
        embed.setImage('attachment://tictactoe.png');
        embed.setTimestamp();
        embed.setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() || '' });

        const attachment = {
            name: 'tictactoe.png',
            attachment: imageBuffer,
        };

        if (this.gameState === TicTacToeGameState.IN_PROGRESS) {
            const buttons = [];
            if (this.getBoard()[0][0] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--0')
                        .setLabel('1')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--0')
                        .setLabel('1')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[0][1] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--1')
                        .setLabel('2')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--1')
                        .setLabel('2')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[0][2] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--2')
                        .setLabel('3')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--2')
                        .setLabel('3')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[1][0] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--3')
                        .setLabel('4')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--3')
                        .setLabel('4')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[1][1] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--4')
                        .setLabel('5')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--4')
                        .setLabel('5')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[1][2] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--5')
                        .setLabel('6')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--5')
                        .setLabel('6')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[2][0] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--6')
                        .setLabel('7')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--6')
                        .setLabel('7')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[2][1] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--7')
                        .setLabel('8')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--7')
                        .setLabel('8')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }
            if (this.getBoard()[2][2] === CellState.EMPTY) {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--8')
                        .setLabel('9')
                        .setStyle(ButtonStyle.Primary)
                );
            } else {
                buttons.push(
                    new ButtonBuilder()
                        .setCustomId('handleTicTacToeButton--8')
                        .setLabel('9')
                        .setStyle(ButtonStyle.Danger)
                        .setDisabled(true)
                );
            }

            const actionRows = [
                new ActionRowBuilder<ButtonBuilder>().addComponents(buttons.slice(0, 3)),
                new ActionRowBuilder<ButtonBuilder>().addComponents(buttons.slice(3, 6)),
                new ActionRowBuilder<ButtonBuilder>().addComponents(buttons.slice(6, 9)),
            ];

            const response = {
                embeds: [embed],
                files: [attachment],
                components: actionRows,
            };

            return response;
        } else {
            const response = {
                embeds: [embed],
                files: [attachment],
                components: [],
            };

            return response;
        }
    }

    private async drawImage(): Promise<Buffer> {
        const size = 256;
        const cellSize = size / 3;
        const image = new Jimp({
            height: size,
            width: size,
            color: 0x000000, // White background
        });
        const lineColor = 0xffffff; // Black
        const lineWidth = 5;
        const font = await loadFont(fontPath);

        // Draw grid lines
        for (let i = 1; i < 3; i++) {
            image.scan(0, i * cellSize - lineWidth / 2, size, lineWidth, (x, y) => {
                image.setPixelColor(lineColor, x, y);
            });
            image.scan(i * cellSize - lineWidth / 2, 0, lineWidth, size, (x, y) => {
                image.setPixelColor(lineColor, x, y);
            });
        }

        // Draw X and O
        for (let row = 0; row < 3; row++) {
            for (let col = 0; col < 3; col++) {
                const cell = this.board[row][col];
                const x = col * cellSize;
                const y = row * cellSize;

                if (cell === CellState.X) {
                    image.print({
                        font,
                        x: x + cellSize / 2 - measureText(font, 'X') / 2,
                        y: y + cellSize / 2 - measureTextHeight(font, 'X', cellSize) / 2,
                        text: 'X',
                    });
                } else if (cell === CellState.O) {
                    image.print({
                        font,
                        x: x + cellSize / 2 - measureText(font, 'O') / 2,
                        y: y + cellSize / 2 - measureTextHeight(font, 'O', cellSize) / 2,
                        text: 'O',
                    });
                }
            }
        }

        // Return the image buffer
        return image.getBuffer(JimpMime.png);
    }
}

export const tictactoeGames = new Map<string, TicTacToeGame>();
