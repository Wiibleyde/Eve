import { changeRadio } from './hope/changeRadio';
import { handleAddRadio } from './hope/handleAddRadio';
import { handleRemoveRadio } from './hope/handleRemoveRadio';
import { handleMotusTry } from './motus/handleMotusTry';
import { handleQuizButton } from './quiz/handleQuizButton';
import { reportQuestionButton } from './quiz/reportQuestionButton';
import {
    addTrackButton,
    backButton,
    iaButton,
    loopButton,
    resumeAndPauseButton,
    skipButton,
} from './music/musicButtons';

export const buttons = {
    handleQuizButton,
    reportQuestionButton,
    changeRadio,
    handleMotusTry,
    handleAddRadio,
    handleRemoveRadio,

    backButton,
    resumeAndPauseButton,
    skipButton,
    loopButton,
    iaButton,
    addTrackButton,
};
