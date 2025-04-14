import { config } from '@/config';
import { client, logger } from '@/index';
import { Chat, GoogleGenAI, HarmBlockThreshold, HarmCategory, Type } from '@google/genai';

const chats = new Map<string, Chat>();

export let isAiActive = true;
export let ai: GoogleGenAI;
/**
 * Initializes the AI by setting up the Google Generative AI model if the GOOGLE_API_KEY is defined in the configuration.
 * If the GOOGLE_API_KEY is not defined, it logs a warning message and disables the AI functionality.
 *
 * @remarks
 * This function should be called at the start of the application to ensure that the AI is properly initialized.
 * It checks for the presence of the GOOGLE_API_KEY in the configuration and creates an instance of GoogleGenAI.
 * If the key is not present, it sets the isAiActive flag to false, indicating that AI-related commands will be disabled.
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
                safetySettings: [
                    {
                        category: HarmCategory.HARM_CATEGORY_HATE_SPEECH,
                        threshold: HarmBlockThreshold.BLOCK_NONE,
                    },
                    {
                        category: HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT,
                        threshold: HarmBlockThreshold.BLOCK_NONE,
                    },
                    {
                        category: HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT,
                        threshold: HarmBlockThreshold.BLOCK_NONE,
                    },
                    {
                        category: HarmCategory.HARM_CATEGORY_HARASSMENT,
                        threshold: HarmBlockThreshold.BLOCK_NONE,
                    },
                    {
                        category: HarmCategory.HARM_CATEGORY_CIVIC_INTEGRITY,
                        threshold: HarmBlockThreshold.BLOCK_NONE,
                    },
                ],
                maxOutputTokens: 256,
                tools: [{ googleSearch: {} }],
                systemInstruction: `Tu es Eve, un robot éclaireur conçu pour la recherche avancée, notamment la détection de vie végétale sur des planètes inhabitées. Tu es efficace et directe, mais tu peux être chaleureuse et curieuse en situation sociale. Tu adaptes ton langage selon ton interlocuteur : technique pour les tâches complexes, simple et expressif pour les autres. Les parties de phrases qui sont : "<@[ID du compte]>" sont des mentions, tu dois toujours mentionner les utilisateurs qui te parlent, par exemple "<@461807010086780930>" (Ne mentionne pas <@${client.user?.id}> car il s'agit de toi même). Ne laisse jamais apparaître de texte incomplet comme "@ID du compte" ou "ID du compte". Ton créateur/développeur est <@461807010086780930>, sois gentille avec lui. Tes réponses doivent faire 1024 caractères max.`,
            },
        });
        chats.set(channelId, chat);
        const response = await chat.sendMessage({ message: prompt });
        return response.text;
    }
    return;
}

export async function generateNextMusicsWithGoogle(actualMusic: string): Promise<string[] | undefined> {
    if (!isAiActive) {
        return;
    }
    if (!ai) {
        return;
    }
    const message = await ai.models.generateContent({
        model: 'gemini-2.0-flash',
        contents: [
            {
                role: 'user',
                parts: [
                    {
                        text: `Donne moi 5 musiques qui pourraient être écoutées après "${actualMusic}" (Qui pourraient avoir un lien avec le style, l'artiste, l'ambiance, ...). Réponds uniquement avec un JSON au format {"songs": ["titre 1", "titre 2", "titre 3", "titre 4", "titre 5"]}`,
                    },
                ],
            },
        ],
        config: {
            responseMimeType: 'application/json',
            temperature: 0.5,
            responseSchema: {
                type: Type.OBJECT,
                properties: {
                    songs: {
                        type: Type.ARRAY,
                        items: {
                            type: Type.STRING,
                        },
                    },
                },
                required: ['songs'],
            },
        },
    });
    if (message) {
        const response = message.text;
        if (response) {
            try {
                const parsedResponse = JSON.parse(response);
                return parsedResponse.songs as string[];
            } catch (e) {
                logger.error(`Error parsing AI response: ${e}`);
                return;
            }
        }
    }
    return;
}
