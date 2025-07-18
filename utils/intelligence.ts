import { Chat, GoogleGenAI, HarmBlockThreshold, HarmCategory } from '@google/genai';
import { config } from './core/config';
import { logger } from '..';
import { client } from '../bot/bot';

const chats = new Map<string, Chat>();

export const iaDisabledServers = [
    '872119977946263632',
    '1152704209192501388',
];

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
 * Creates a new chat instance for a channel
 *
 * @param channelId - The unique identifier for the chat channel
 * @returns A new chat instance
 */
function createChatForChannel(channelId: string): Chat {
    const chat = ai.chats.create({
        model: 'gemini-2.5-flash',
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
            tools: [{ googleSearch: {} }],
            systemInstruction: `(Tu es connecté à un salon Discord donc tu parles au milieu des personnes qui te le demandent et tu interviens comme un compte Discord classique.) Tu es Eve, un robot éclaireur conçu pour la recherche avancée, notamment la détection de vie végétale sur des planètes inhabitées. Tu es efficace et directe, mais tu peux être chaleureuse et curieuse en situation sociale. Tu adaptes ton langage selon ton interlocuteur : technique pour les tâches complexes, simple et expressif pour les autres. Les parties de phrases qui sont : "<@[ID du compte]>" sont des mentions, tu dois toujours mentionner les utilisateurs qui te parlent, par exemple "<@461807010086780930>" (Ne mentionne pas <@${client.user?.id}> car il s'agit de toi même). Ne laisse jamais apparaître de texte incomplet comme "@ID du compte" ou "ID du compte". Ton créateur/développeur est <@461807010086780930>, sois gentille avec lui, ping le seulement quand il te parle. Tes réponses doivent faire 1024 caractères max et doivent être courtes et consices. Et aussi, voici les emojis que tu peux utiliser (en plus de ceux de base): <:eve:1310908312144969760>, <:eveHappy:1310908313457659934>, <:eveLove:1310908314820935690> <:eveNervous:1310908316980875285> <:eveSleep:1310908318247813121> <:eveSurprised:1310908319958958120> <:eveUnhappy:1310908322009976883>`,
        },
    });

    chats.set(channelId, chat);
    return chat;
}

/**
 * Generates a response using Google's AI model based on the provided prompt.
 *
 * @param channelId - The unique identifier for the chat channel.
 * @param prompt - The prompt or question asked by the user.
 * @param userAsking - The identifier of the user asking the question.
 * @returns A promise that resolves to the generated response as a string, or null if there's an error.
 *
 * @throws Will not throw errors, but will return null and log errors internally.
 */
export async function generateWithGoogle(
    channelId: string,
    prompt: string,
    userAsking: string
): Promise<string | null> {
    try {
        // Verify AI is active and initialized
        if (!isAiActive) {
            logger.warn("L'IA est désactivée. Impossible de générer une réponse.");
            return null;
        }

        if (!ai) {
            logger.error("L'IA n'a pas été initialisée. Appelez d'abord initAi().");
            return null;
        }

        // Validate parameters
        if (!channelId || !prompt || !userAsking) {
            logger.error('Paramètres invalides pour generateWithGoogle. Tous les paramètres sont requis.');
            return null;
        }

        // Get or create chat
        let chat: Chat;
        if (chats.has(channelId)) {
            const existingChat = chats.get(channelId);
            if (!existingChat) {
                logger.error(`Chat trouvé dans la map mais undefined pour channelId: ${channelId}`);
                return null;
            }
            chat = existingChat;
        } else {
            chat = createChatForChannel(channelId);
        }

        // Format message and send
        const formattedMessage = `<@${userAsking}> écrit : ${prompt}`;
        const response = await chat.sendMessage({ message: formattedMessage });

        if (!response || !response.text) {
            logger.warn(`Aucune réponse générée par l'IA pour le prompt: "${prompt.substring(0, 50)}..."`);
            return null;
        }

        return response.text;
    } catch (error) {
        logger.error(
            `Erreur lors de la génération de la réponse: ${error instanceof Error ? error.message : String(error)}`
        );
        return null;
    }
}
