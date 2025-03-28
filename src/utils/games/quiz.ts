import { logger } from '@/index';
import { prisma } from '../database';

const quizApiUrl = 'https://quizzapi.jomoreschi.fr/api/v1/quiz?limit=50';

/**
 * Fetches a quiz question from an external API and inserts it into the database.
 *
 * This function performs the following steps:
 * 1. Fetches quiz data from a predefined API URL.
 * 2. Parses the response to extract the first quiz question.
 * 3. Constructs a `QuizType` object from the extracted data.
 * 4. Attempts to insert the quiz question into the database using Prisma.
 *
 * If a unique constraint violation occurs (error code "P2002"), the function will silently return.
 * Any other errors encountered during the database insertion will be logged.
 *
 * @returns {Promise<void>} A promise that resolves when the operation is complete.
 *
 * @throws {Error} Throws an error if the fetch operation fails or if an unexpected error occurs during database insertion.
 */
export async function insertQuestionInDB(): Promise<void> {
    const response = await fetch(quizApiUrl);
    const data = await response.json();
    const quizzesJson = data.quizzes;

    for (const quizJson of quizzesJson) {
        const quiz: QuizType = {
            question: quizJson.question,
            answer: quizJson.answer,
            badAnswers: quizJson.badAnswers,
            category: quizJson.category,
            difficulty: quizJson.difficulty,
            createdAt: Date.now(),
        };
        try {
            await prisma.quizQuestions.create({
                data: {
                    question: quiz.question,
                    answer: quiz.answer,
                    badAnswer1: quiz.badAnswers[0],
                    badAnswer2: quiz.badAnswers[1],
                    badAnswer3: quiz.badAnswers[2],
                    category: quiz.category,
                    difficulty: quiz.difficulty,
                    guildId: '0',
                },
            });
            logger.debug(`Question de quiz insérée dans la base de données : ${quiz.question} - ${quiz.answer}`)
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: Error | any) {
            if (error.code === 'P2002') {
                return;
            } else {
                logger.error(
                    `Erreur lors de l'insertion de la question de quiz dans la base de données : ${error.message}`
                );
            }
        }
    }
}
