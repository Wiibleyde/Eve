import type { ButtonInteraction } from 'discord.js';
import { jokeSetPublicButton } from './handlers/jokeSetPublic';
import { handleMotusTry } from './handlers/game/handleMotusTry';
import { handleQuizButton } from './handlers/game/handleQuizButton';
import { handleLsmsDuty } from './handlers/rp/handleLsmsDuty';
import { handleLsmsOnCall } from './handlers/rp/handleLsmsOnCall';
import { handleLsmsRadioAdd } from './handlers/rp/handleLsmsRadioAdd';
import { handleLsmsRadioRemove } from './handlers/rp/handleLsmsRadioRemove';
import { handleLsmsRadioEdit } from './handlers/rp/handleLsmsRadioEdit';
import { laboCancelButton } from './handlers/rp/laboCancelButton';
import { lotoBuy } from './handlers/rp/lotoBuy';
import { lotoDraw } from './handlers/rp/lotoDraw';
import { removeTickets } from './handlers/rp/removeTickets';
import { lotoEditPlayer } from './handlers/rp/lotoEditPlayer';

export const buttons: Record<string, (interaction: ButtonInteraction) => Promise<void>> = {
    jokeSetPublicButton: jokeSetPublicButton,
    handleMotusTry: handleMotusTry,
    handleQuizButton: handleQuizButton,
    handleLsmsDuty: handleLsmsDuty,
    handleLsmsOnCall: handleLsmsOnCall,
    lsmsRadioAdd: handleLsmsRadioAdd,
    lsmsRadioRemove: handleLsmsRadioRemove,
    lsmsRadioEdit: handleLsmsRadioEdit,
    laboCancelButton: laboCancelButton,
    lotoBuy: lotoBuy,
    lotoDraw: lotoDraw,
    removeTickets: removeTickets,
    lotoEditPlayer: lotoEditPlayer,
};
