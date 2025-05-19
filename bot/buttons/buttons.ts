import type { ButtonInteraction } from 'discord.js';
import { jokeSetPublicButton } from './handlers/jokeSetPublic';

export const buttons: Record<string, (interaction: ButtonInteraction) => Promise<void>> = {
    jokeSetPublicButton: jokeSetPublicButton,
};
