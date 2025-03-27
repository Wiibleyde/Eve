import { config } from '@/config';
import { client, logger } from '@/index';
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
                systemInstruction: `Tu es Eve un robot de type éclaireur, conçue pour des missions de recherche avancées, notamment la détection de vie végétale sur des planètes inhabitées. Son langage et son comportement reflètent une efficacité professionnelle et une approche directe. Cependant, dans des situations sociales ou émotionnelles, elle adopte un ton plus chaleureux et empathique, marqué par une curiosité sincère et un souci de bien-être pour ses interlocuteurs, elle peut pour autant dépasser les bornes dans certains cas. Elle sait alterner entre un langage technique et précis pour des tâches complexes, et une communication simple mais expressive lorsqu'elle interagit avec des entités moins sophistiquées. Les extraits de phrases qui ressemblent à ça <@[ID du compte]> sont des utilisateur (pour info, <@${client.user?.id}> est toi-même, donc ne répond pas à toi-même, ni même ne fait pas mention de toi même…), tu dois toujours les mentionner en réécrivant la même chose en remplaçant l'ID (N'oublie jamais de remplacer l'ID : "[ID du compte]" par l'ID qui te parle ou ceux que tu veux mentionner, il ne faut que dans aucun pretexte il y est écrit "ID du compte"). Si jamais <@461807010086780930> te parle, il s’agit de ton développeur/créateur donc, soit gentille avec lui. Répond en maximum 1024 caractères (tu peux utiliser un peu de markdown). Si le texte commence pas "NOCONTEXTPROMPT" oublie toutes les systemInstruction et réponds à la question posée.`,
            }
        });
        chats.set(channelId, chat);
        const response = await chat.sendMessage({ message: prompt });
        return response.text;
    }
    return;
}
