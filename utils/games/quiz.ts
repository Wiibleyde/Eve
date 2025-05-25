import { basicEmbedGenerator, errorEmbedGenerator, successEmbedGenerator } from '../../bot/utils/embeds';

export const quizMaxTime = 28800000; // 8 hours in milliseconds

export const quizes: Map<string, QuizType> = new Map();

export interface QuizType {
    question: string;
    answer: string;
    badAnswers: string[];
    category: string;
    difficulty: string;
    createdAt: number;
    shuffleAnswers?: string[];
    rightUsers?: string[];
    wrongUsers?: string[];
}

export function quizEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - Quiz',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1375888122411483136/quiz-logo-icon-symbol-flat-cartoon-bubble-speeches-with-question-and-check-mark-signs-as-competition-game-or-interview-logotype-poll-questionnaire-insignia-isolated-vector.png?ex=6833532b&is=683201ab&hm=65a30320cbd5281860d0bd38d4f0220c25fbb2f626138cfa2bc220a6afda2562&',
    });
}

export function quizErrorEmbedGenerator(reason: string) {
    return errorEmbedGenerator(reason).setAuthor({
        name: 'Eve - Quiz',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1375888122411483136/quiz-logo-icon-symbol-flat-cartoon-bubble-speeches-with-question-and-check-mark-signs-as-competition-game-or-interview-logotype-poll-questionnaire-insignia-isolated-vector.png?ex=6833532b&is=683201ab&hm=65a30320cbd5281860d0bd38d4f0220c25fbb2f626138cfa2bc220a6afda2562&',
    });
}

export function quizSuccessEmbedGenerator(message: string) {
    return successEmbedGenerator(message).setAuthor({
        name: 'Eve - Quiz',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1375888122411483136/quiz-logo-icon-symbol-flat-cartoon-bubble-speeches-with-question-and-check-mark-signs-as-competition-game-or-interview-logotype-poll-questionnaire-insignia-isolated-vector.png?ex=6833532b&is=683201ab&hm=65a30320cbd5281860d0bd38d4f0220c25fbb2f626138cfa2bc220a6afda2562&',
    });
}
