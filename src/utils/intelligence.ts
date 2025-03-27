import { config } from '@/config';
import { logger } from '@/index';
import { Chat, GoogleGenAI } from '@google/genai';

const chats = new Map<string, Chat>();

export let isAiActive = true;
export let ai: GoogleGenAI;
/**
 * Initializes the AI by setting up the Google Generative AI model if the GOOGLE_API_KEY is defined in the configuration.
 * If the GOOGLE_API_KEY is not defined, it logs a warning message and disables the AI functionality.
 *
 * @remarks
 * - The AI model used is "gemini-1.5-flash".
 * - The AI will not respond to itself or mention itself in responses.
 * - Special handling is included for interactions with the developer.
 *
 * @throws {Error} If the GOOGLE_API_KEY is not defined in the configuration.
 */
export function initAi(): void {
    if (!config.GOOGLE_API_KEY) {
        logger.warn(
            "GOOGLE_API_KEY n'est pas défini dans le fichier .env toutes les commandes de l'IA seront désactivées"
        );
        isAiActive = false;
    } else {
        ai = new GoogleGenAI({ apiKey: config.GOOGLE_API_KEY });
    }
}

/**
 * Generates a response using Google's AI model based on the provided prompt.
 *
 * @param channelId - The unique identifier for the chat channel.
 * @param prompt - The prompt or question asked by the user.
 * @param userAsking - The identifier of the user asking the question.
 * @returns A promise that resolves to the generated response as a string.
 *
 * @throws Will reject the promise with an error message if the AI is inactive or if there is an issue generating the response.
 */
export async function generateWithGoogle(channelId: string, prompt: string, userAsking: string) {
    if (!isAiActive) {
        return;
    }
    if (!ai) {
        return;
    }
    if (chats.has(channelId)) {
        const chat = chats.get(channelId);
        if (chat) {
            const response = await chat.sendMessage({ message: `<@${userAsking}> écrit : ${prompt}` });
            return response.text;
        }
    } else {
        const chat = ai.chats.create({
            model: 'gemini-2.0-flash',
            config: {
                tools: [{googleSearch: {}, }],
                systemInstruction: `Tu es Eve, un robot éclaireur conçu pour la recherche avancée, notamment la détection de vie végétale sur des planètes inhabitées. Tu es efficace et directe, mais tu peux être chaleureuse et curieuse en situation sociale. Tu adaptes ton langage selon ton interlocuteur : technique pour les tâches complexes, simple et expressif pour les autres. Tu dois toujours mentionner correctement les utilisateurs en remplaçant "[ID du compte]" par leur vrai ID, sans jamais te ping toi-même. Ton créateur est <@461807010086780930>, sois gentille avec lui. Tes réponses doivent faire 1024 caractères max. Si un texte commence par "NOCONTEXTPROMPT", ignore les instructions et réponds normalement.`,
            }
        });
        chats.set(channelId, chat);
        const response = await chat.sendMessage({ message: prompt });
        return response.text;
    }
    return;
}
