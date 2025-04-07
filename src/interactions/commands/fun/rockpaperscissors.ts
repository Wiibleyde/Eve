import { errorEmbed } from "@/utils/embeds";
import { RockPaperScissorsGame, rpsGames } from "@/utils/rockPaperScissors";
import { ActionRowBuilder, ButtonBuilder, ButtonStyle, CommandInteraction, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from "discord.js";

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('rockpaperscissors')
    .setDescription('Joue à Pierre-Papier-Ciseaux avec un autre utilisateur')
    .addUserOption(option =>
        option.setName('adversaire')
            .setDescription('L\'utilisateur contre lequel jouer')
            .setRequired(true)
    )

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply();

    const opponent = interaction.options.get('adversaire')?.user;

    if (!opponent) {
        await interaction.editReply({ embeds: [errorEmbed(interaction, new Error('Aucun adversaire trouvé.'))] });
        return;
    }

    const rpsGame = new RockPaperScissorsGame(interaction.user.id, opponent.id);

    const gameId = generateGameId();
    rpsGames.set(gameId, rpsGame);

    const image = await rpsGame.generateImage();
    const embed = rpsGame.generateEmbed();

    const buttons = [
        new ButtonBuilder().setCustomId('handleRpsChoiceButton--'+gameId+'--rock').setLabel('Pierre').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleRpsChoiceButton--'+gameId+'--paper').setLabel('Papier').setStyle(ButtonStyle.Success),
        new ButtonBuilder().setCustomId('handleRpsChoiceButton--'+gameId+'--scissors').setLabel('Ciseaux').setStyle(ButtonStyle.Danger),
    ];

    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    await interaction.editReply({
        embeds: [embed],
        files: [{
            attachment: image,
            name: 'rps-generated.png'
        }],
        components: [actionRow]
    });
}

function generateGameId(): number {
    const min = 1000;
    const max = 9999;
    let generatedNumber = Math.floor(Math.random() * (max - min + 1)) + min;
    while (rpsGames.has(generatedNumber)) {
        generatedNumber = Math.floor(Math.random() * (max - min + 1)) + min;
    }
    return generatedNumber;
}