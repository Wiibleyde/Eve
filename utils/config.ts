import dotenv from 'dotenv';

dotenv.config();

const {
    DISCORD_TOKEN,
    DISCORD_CLIENT_ID,
    EVE_HOME_GUILD,
    OWNER_ID,
    LOGS_WEBHOOK_URL,
    GOOGLE_API_KEY,
    REPORT_CHANNEL,
    MP_CHANNEL,
    BLAGUE_API_TOKEN,
    NASA_API_KEY,
    TWITCH_CLIENT_ID,
    TWITCH_CLIENT_SECRET,
} = process.env;

const { MUSIC_MODULE: MUSIC_MODULE_STRING } = process.env;

if (
    !DISCORD_TOKEN ||
    !DISCORD_CLIENT_ID ||
    !EVE_HOME_GUILD ||
    !OWNER_ID ||
    !LOGS_WEBHOOK_URL ||
    !GOOGLE_API_KEY ||
    !REPORT_CHANNEL ||
    !MP_CHANNEL ||
    !BLAGUE_API_TOKEN ||
    !NASA_API_KEY ||
    !TWITCH_CLIENT_ID ||
    !TWITCH_CLIENT_SECRET
) {
    const missingVars = [];
    if (!DISCORD_TOKEN) missingVars.push('DISCORD_TOKEN');
    if (!DISCORD_CLIENT_ID) missingVars.push('DISCORD_CLIENT_ID');
    if (!EVE_HOME_GUILD) missingVars.push('EVE_HOME_GUILD');
    if (!OWNER_ID) missingVars.push('OWNER_ID');
    if (!LOGS_WEBHOOK_URL) missingVars.push('LOGS_WEBHOOK_URL');
    if (!GOOGLE_API_KEY) missingVars.push('GOOGLE_API_KEY');
    if (!REPORT_CHANNEL) missingVars.push('REPORT_CHANNEL');
    if (!MP_CHANNEL) missingVars.push('MP_CHANNEL');
    if (!BLAGUE_API_TOKEN) missingVars.push('BLAGUE_API_TOKEN');
    if (!NASA_API_KEY) missingVars.push('NASA_API_KEY');
    if (!TWITCH_CLIENT_ID) missingVars.push('TWITCH_CLIENT_ID');
    if (!TWITCH_CLIENT_SECRET) missingVars.push('TWITCH_CLIENT_SECRET');
    throw new Error(`Missing environment variables: ${missingVars.join(', ')}`);
}

const MUSIC_MODULE = MUSIC_MODULE_STRING?.toLowerCase() === 'true' ? true : false;

/**
 * Configuration object for the EVe Assistant application.
 *
 * @property {string} DISCORD_TOKEN - The token used to authenticate with the Discord API.
 * @property {string} DISCORD_CLIENT_ID - The client ID of the Discord application.
 * @property {string} EVE_HOME_GUILD - The ID of the home guild (server) for the Eve Assistant.
 * @property {string} OWNER_ID - The ID of the owner of the bot.
 * @property {string} LOGS_WEBHOOK_URL - The URL of the webhook used for logging.
 * @property {string} GOOGLE_API_KEY - The API key used to authenticate with Google services.
 * @property {string} REPORT_CHANNEL - The ID of the channel where reports are sent.
 * @property {string} MP_CHANNEL - The ID of the channel where MPs are sent.
 * @property {string} BLAGUE_API_TOKEN - The token used to authenticate with the Blague API.
 * @property {string} NASA_API_KEY - The API key used to authenticate with the NASA API.
 * @property {boolean} MUSIC_MODULE - Flag indicating whether the music module is enabled.
 * @property {string} TWITCH_CLIENT_ID - The client ID for Twitch API authentication.
 * @property {string} TWITCH_CLIENT_SECRET - The client secret for Twitch API authentication.
 */
export const config = {
    DISCORD_TOKEN,
    DISCORD_CLIENT_ID,
    EVE_HOME_GUILD,
    OWNER_ID,
    LOGS_WEBHOOK_URL,
    GOOGLE_API_KEY,
    REPORT_CHANNEL,
    MP_CHANNEL,
    BLAGUE_API_TOKEN,
    NASA_API_KEY,
    MUSIC_MODULE,
    TWITCH_CLIENT_ID,
    TWITCH_CLIENT_SECRET,
};