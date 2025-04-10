import { changeRadio } from './rp/changeRadio';
import { handleAddRadio } from './rp/handleAddRadio';
import { handleRemoveRadio } from './rp/handleRemoveRadio';
import { handleMotusTry } from './motus/handleMotusTry';
import { handleQuizButton } from './quiz/handleQuizButton';
import { reportQuestionButton } from './quiz/reportQuestionButton';
import { addIaSuggestion, backButton, iaButton, loopButton, resumeAndPauseButton, skipButton } from './music/musicButtons';
import { handleTicTacToeButton } from './tictactoe/handleTicTacToeButton';

export const buttons = {
    handleQuizButton,
    reportQuestionButton,
    changeRadio,
    handleMotusTry,
    handleAddRadio,
    handleRemoveRadio,
    handleTicTacToeButton,

    backButton,
    resumeAndPauseButton,
    skipButton,
    loopButton,
    iaButton,
    addIaSuggestion,
};
