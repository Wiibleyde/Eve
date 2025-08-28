import type { ButtonInteraction } from 'discord.js';
import { jokeSetPublicButton } from './handlers/jokeSetPublic';
import { handleMotusTry } from './handlers/game/handleMotusTry';
import { handleQuizButton } from './handlers/game/handleQuizButton';
import { handleLsmsDuty } from './handlers/rp/handleLsmsDuty';
import { handleLsmsOnCall } from './handlers/rp/handleLsmsOnCall';
import { laboCancelButton } from './handlers/rp/laboCancelButton';

export const buttons: Record<string, (interaction: ButtonInteraction) => Promise<void>> = {
    jokeSetPublicButton: jokeSetPublicButton,
    handleMotusTry: handleMotusTry,
    handleQuizButton: handleQuizButton,
    handleLsmsDuty: handleLsmsDuty,
    handleLsmsOnCall: handleLsmsOnCall,
    laboCancelButton: laboCancelButton,
};
