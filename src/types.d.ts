interface QuizType {
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
