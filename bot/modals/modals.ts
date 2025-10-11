import type { ModalSubmitInteraction } from 'discord.js';
import { handleMotusTryModal } from './handlers/game/handleMotusTryModal';
import { lotoBuyModal } from './handlers/rp/lotoBuyModal';
import { removeTicketsModal } from './handlers/rp/removeTicketsModal';
import { lotoEditPlayerModal } from './handlers/rp/lotoEditPlayerModal';

export const modals: Record<string, (interaction: ModalSubmitInteraction) => Promise<void>> = {
    handleMotusTryModal: handleMotusTryModal,
    lotoBuyModal: lotoBuyModal,
    removeTicketsModal: removeTicketsModal,
    lotoEditPlayerModal: lotoEditPlayerModal,
};
