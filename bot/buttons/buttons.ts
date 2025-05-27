import type { ButtonInteraction } from 'discord.js';
import { jokeSetPublicButton } from './handlers/jokeSetPublic';
import { handleMotusTry } from './handlers/handleMotusTry';
import { handleQuizButton } from './handlers/handleQuizButton';
import { handleLsmsDuty } from './handlers/handleLsmsDuty';
import { handleLsmsOnCall } from './handlers/handleLsmsOnCall';

export const buttons: Record<string, (interaction: ButtonInteraction) => Promise<void>> = {
    jokeSetPublicButton: jokeSetPublicButton,
    handleMotusTry: handleMotusTry,
    handleQuizButton: handleQuizButton,
    handleLsmsDuty: handleLsmsDuty,
    handleLsmsOnCall: handleLsmsOnCall,
};
