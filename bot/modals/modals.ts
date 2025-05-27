import type { ModalSubmitInteraction } from 'discord.js';
import { handleMotusTryModal } from './handlers/handleMotusTryModal';

export const modals: Record<string, (interaction: ModalSubmitInteraction) => Promise<void>> = {
    handleMotusTryModal: handleMotusTryModal,
};
