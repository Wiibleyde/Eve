import { prisma } from '@/utils/database';
import { errorEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import {
    CommandInteraction,
    SlashCommandBuilder,
    EmbedBuilder,
    ButtonBuilder,
    ButtonStyle,
    ActionRowBuilder,
    TextChannel,
} from 'discord.js';
import { maxTime, quizes } from '@/interactions/buttons/quiz/handleQuizButton';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('quiz')
    .setDescription('Affiche une question de quiz aléatoire');

/**
 * Executes the quiz command, fetching a random quiz question from the database and sending it as an embedded message
 * with interactive buttons for the user to select an answer.
 *
 * @param {CommandInteraction} interaction - The interaction object representing the command invocation.
 * @returns {Promise<void>} A promise that resolves when the command execution is complete.
 *
 * @throws {Error} If no quiz questions are found in the database.
 * @throws {Error} If an error occurs while retrieving the quiz question.
 * @throws {Error} If an error occurs while sending the quiz question to the channel.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    const questionCount = await prisma.quizQuestions.count();
    if (questionCount === 0) {
        await interaction.reply({
            embeds: [
                errorEmbed(interaction, new Error("Aucune question de quiz n'a été trouvée dans la base de données.")),
            ],
            ephemeral: true,
        });
        return;
    }
    const randomQuiz = await prisma.quizQuestions.findMany({
        take: 1,
        skip: Math.floor(Math.random() * questionCount),
        include: {
            author: true,
        },
    });
    const quizJson = randomQuiz[0];
    const quiz: QuizType = {
        question: quizJson.question,
        answer: quizJson.answer,
        badAnswers: [quizJson.badAnswer1, quizJson.badAnswer2, quizJson.badAnswer3],
        category: quizJson.category,
        difficulty: quizJson.difficulty,
        createdAt: Date.now(),
    };

    const invalidQuizTimestamp = quiz.createdAt + maxTime;

    if (!quiz) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error('Une erreur est survenue lors de la récupération de la question de quiz.')
                ),
            ],
            ephemeral: true,
        });
        return;
    }

    const shuffledAnswers = [quiz.answer, ...quiz.badAnswers].sort(() => Math.random() - 0.5);
    quiz.shuffleAnswers = shuffledAnswers;

    const formattedCategory = quiz.category.charAt(0).toUpperCase() + quiz.category.slice(1).replace(/_/g, ' ');

    const embed = new EmbedBuilder()
        .setTitle('Question de quiz')
        .setDescription(
            `${'```'}${quiz.question}${'```'}${backSpace}1) ${'`'}${shuffledAnswers[0]}${'`'}${backSpace}2) ${'`'}${shuffledAnswers[1]}${'`'}${backSpace}3) ${'`'}${shuffledAnswers[2]}${'`'}${backSpace}4) ${'`'}${shuffledAnswers[3]}${'`'}`
        )
        .addFields(
            { name: 'Catégorie / difficulté', value: `${formattedCategory} / ${quiz.difficulty}`, inline: true },
            { name: 'Invalide', value: `<t:${Math.floor(invalidQuizTimestamp / 1000)}:R>`, inline: true }
        )
        .setColor(0x4b0082)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    if (quizJson.author) {
        embed.addFields({ name: 'Auteur', value: `<@${quizJson.author.userId}>`, inline: true });
    } else {
        embed.addFields({ name: 'Demandé par', value: `<@${interaction.user.id}>`, inline: true });
    }

    const buttons = [
        new ButtonBuilder().setCustomId('handleQuizButton--1').setLabel('1').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--2').setLabel('2').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--3').setLabel('3').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--4').setLabel('4').setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('reportQuestionButton')
            .setLabel('Signaler une erreur')
            .setStyle(ButtonStyle.Danger),
    ];

    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    const channel = interaction.channel as TextChannel;
    if (!channel) {
        await interaction.reply({
            embeds: [
                errorEmbed(interaction, new Error("Une erreur est survenue lors de l'envoi de la question de quiz.")),
            ],
            ephemeral: true,
        });
        return;
    }

    const messageResponse = await channel.send({ embeds: [embed], components: [actionRow] });

    quizes.set(messageResponse.id, quiz);

    await interaction.reply({ content: 'Question de quiz envoyée !', ephemeral: true });

    await prisma.quizQuestions.update({
        where: {
            uuid: quizJson.uuid,
        },
        data: {
            lastTimeUsed: new Date(),
        },
    });
}

export function isMessageQuizQuestion(messageId: string) {
    return quizes.has(messageId);
}
