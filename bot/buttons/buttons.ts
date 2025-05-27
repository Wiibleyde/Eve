import type { ButtonInteraction } from 'discord.js';
import { jokeSetPublicButton } from './handlers/jokeSetPublic';
import { handleMotusTry } from './handlers/handleMotusTry';
import { handleQuizButton } from './handlers/handleQuizButton';

export const buttons: Record<string, (interaction: ButtonInteraction) => Promise<void>> = {
    jokeSetPublicButton: jokeSetPublicButton,
    handleMotusTry: handleMotusTry,
    handleQuizButton: handleQuizButton,
};
