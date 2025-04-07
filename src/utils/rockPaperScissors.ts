import { Jimp, JimpMime, loadFont, ResizeStrategy } from "jimp";
import { client } from "..";
import { EmbedBuilder } from "discord.js";

enum GameStatus {
    PLAYING = 'playing',
    END = 'end',
}

enum PlayerChoice {
    ROCK,
    PAPER,
    SCISSORS,
}

const sans16WhitePath = 'assets/fonts/open-sans/open-sans-16-white/open-sans-16-white.fnt';

const background = 'assets/img/quote.png';

const paperImage = 'assets/img/rps/paper.png';
const rockImage = 'assets/img/rps/rock.png';
const scissorsImage = 'assets/img/rps/scissors.png';


export class RockPaperScissorsGame {
    private player1: string;
    private player2: string;
    private player1Choice: PlayerChoice | null;
    private player2Choice: PlayerChoice | null;
    private status: GameStatus;
    private winner: string | null = null;

    constructor(player1: string, player2: string) {
        this.player1 = player1;
        this.player2 = player2;
        this.player1Choice = null;
        this.player2Choice = null;
        this.status = GameStatus.PLAYING;
    }

    public recordChoice(player: string, choice: PlayerChoice): void {
        if (this.status === GameStatus.END) {
            throw new Error('Game has already ended.');
        }

        if (player === this.player1) {
            this.player1Choice = choice;
        } else if (player === this.player2) {
            this.player2Choice = choice;
        } else {
            throw new Error('Invalid player.');
        }

        if (this.player1Choice !== null && this.player2Choice !== null) {
            this.status = GameStatus.END;
        }

        if (this.player1Choice !== null && this.player2Choice !== null) {
            this.determineWinner();
        }
    }

    private determineWinner(): void {
        if (this.player1Choice === this.player2Choice) {
            this.winner = null; // Tie
        } else if (
            (this.player1Choice === PlayerChoice.ROCK && this.player2Choice === PlayerChoice.SCISSORS) ||
            (this.player1Choice === PlayerChoice.PAPER && this.player2Choice === PlayerChoice.ROCK) ||
            (this.player1Choice === PlayerChoice.SCISSORS && this.player2Choice === PlayerChoice.PAPER)
        ) {
            this.winner = this.player1;
        } else {
            this.winner = this.player2;
        }
    }

    public getWinner(): string | null {
        return this.winner;
    }

    public getStatus(): GameStatus {
        return this.status;
    }

    public getChoices(): { player1Choice: PlayerChoice | null; player2Choice: PlayerChoice | null } {
        return {
            player1Choice: this.player1Choice,
            player2Choice: this.player2Choice,
        };
    }

    public async generateImage(): Promise<Buffer> {
        const image = await Jimp.read(background);
        const sans16WhiteFont = await loadFont(sans16WhitePath);

        const player1Image = (await client.users.fetch(this.player1)).avatarURL({
            extension: 'png',
            size: 512,
        });

        const player2Image = (await client.users.fetch(this.player2)).avatarURL({
            extension: 'png',
            size: 512,
        });

        if (player1Image) {
            const maxPictureHeight = image.bitmap.height - 18;
            const profilePicture1 = await Jimp.read(player1Image);
            profilePicture1.opacity(0.3);
            image.composite(
                profilePicture1.resize({ h: maxPictureHeight, mode: ResizeStrategy.BEZIER }), 
                11,
                9
            );
        }

        if (player2Image) {
            const maxPictureHeight = image.bitmap.height - 18;
            const profilePicture2 = await Jimp.read(player2Image);
            profilePicture2.opacity(0.3);
            const profileWidth = (maxPictureHeight * profilePicture2.bitmap.width) / profilePicture2.bitmap.height;
            image.composite(
                profilePicture2.resize({ h: maxPictureHeight, mode: ResizeStrategy.BEZIER }), 
                image.bitmap.width - profileWidth - 11,
                9
            );
        }

        const player1ChoiceImage = await this.getChoiceImage(this.player1Choice);
        const player2ChoiceImage = await this.getChoiceImage(this.player2Choice);
        if (player1ChoiceImage && player2ChoiceImage) {
            const choiceImage1 = await Jimp.read(player1ChoiceImage);
            image.composite(choiceImage1.resize({ w: 200, h: 200 }), 100, 300);

            const choiceImage2 = await Jimp.read(player2ChoiceImage);
            image.composite(choiceImage2.resize({ w: 200, h: 200 }), image.bitmap.width - 300, 300);
        }

        const player1XPosition = 50;
        image.print({
            x: player1XPosition,
            y: 550,
            font: sans16WhiteFont,
            text: this.player1,
        });

        const player2XPosition = image.bitmap.width - 150;
        image.print({
            x: player2XPosition,
            y: 550,
            font: sans16WhiteFont,
            text: this.player2,
        });

        const buffer = await image.getBuffer(JimpMime.png);
        return buffer;
    }

    private async getChoiceImage(choice: PlayerChoice | null): Promise<string | null> {
        if (choice === PlayerChoice.ROCK) {
            return rockImage;
        } else if (choice === PlayerChoice.PAPER) {
            return paperImage;
        } else if (choice === PlayerChoice.SCISSORS) {
            return scissorsImage;
        }
        return null;
    }

    public generateEmbed(): EmbedBuilder {
        const embed = new EmbedBuilder()
            .setTitle('Pierre, papier, ciseaux !')
            .setDescription(`<@${this.player1}> vs <@${this.player2}>`)
            .setColor(0x4b0082)
            .setTimestamp()
            .setImage('attachment://rps-generated.png')
            .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: client.user?.displayAvatarURL() });

        if (this.status === GameStatus.END) {
            if (this.winner) {
                embed.setDescription(`Le gagnant est <@${this.winner}> !`);
            } else {
                embed.setDescription('C\'est une égalité !');
            }
        }
        if (this.player1Choice !== null && this.player2Choice !== null) {
            embed.addFields(
                { name: `<@${this.player1}> a choisi`, value: this.player1Choice.toString(), inline: true },
                { name: `<@${this.player2}> a choisi`, value: this.player2Choice.toString(), inline: true }
            );
        }
        return embed;
    }
}


export const rpsGames: Map<number, RockPaperScissorsGame> = new Map();