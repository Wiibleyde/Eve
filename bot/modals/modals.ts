import type { ModalSubmitInteraction } from 'discord.js';
import { handleMotusTryModal } from './handlers/game/handleMotusTryModal';
import { lotoBuyModal } from './handlers/rp/lotoBuyModal';
import { removeTicketsModal } from './handlers/rp/removeTicketsModal';
import { lotoEditPlayerModal } from './handlers/rp/lotoEditPlayerModal';
import { lsmsRadioAddModal } from './handlers/rp/lsmsRadioAddModal';
import { lsmsRadioEditModal } from './handlers/rp/lsmsRadioEditModal';

export const modals: Record<string, (interaction: ModalSubmitInteraction) => Promise<void>> = {
    handleMotusTryModal: handleMotusTryModal,
    lotoBuyModal: lotoBuyModal,
    removeTicketsModal: removeTicketsModal,
    lotoEditPlayerModal: lotoEditPlayerModal,
    lsmsRadioAddModal: lsmsRadioAddModal,
    lsmsRadioEditModal: lsmsRadioEditModal,
};
