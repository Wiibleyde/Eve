import * as chill from './fun/chill';
import * as ping from './general/ping';
import * as info from './general/info';
import * as talk from './general/talk';
import * as quote from './general/quote';
import * as birthday from './general/birthday';
import * as help from './general/help';

import * as config from './config/config';
import * as rename from './config/rename';
import * as addstreamer from './config/addstreamer';
import * as removestreamer from './config/removestreamer';

import * as cat from './fun/cat';
import * as dog from './fun/dog';
import * as blague from './fun/blague';
import * as apod from './fun/apod';

import * as quiz from './fun/quiz/quiz';
import * as addquestion from './fun/quiz/addquestion';
import * as quizstats from './fun/quiz/quizstats';
import * as leaderboard from './fun/quiz/leaderboard';

import * as motus from './fun/motus';

import * as createradio from './hope/createradio';

import * as createcalendar from './calendar/createcalendar';

import * as chaban from './utils/chaban';

import * as logs from './dev/logs';
import * as debug from './dev/debug';
import * as maintenance from './dev/maintenance';

import * as back from './music/back';
import * as clear from './music/clear';
import * as filter from './music/filter';
import * as pause from './music/pause';
import * as play from './music/play';
import * as resume from './music/resume';
import * as skip from './music/skip';
import * as loop from './music/loop';
import * as syncedlyrics from './music/syncedlyrics';
import * as queue from './music/queue';
import * as nowplaying from './music/nowplaying';
import * as remove from './music/remove';
import * as stop from './music/stop';

export const commands = {
    ping,
    info,
    talk,
    birthday,
    config,
    rename,
    quote,
    cat,
    dog,
    quiz,
    help,
    addquestion,
    quizstats,
    leaderboard,
    blague,
    chill,
    chaban,
    apod,
    motus,
    addstreamer,
    removestreamer,

    // Calendar commands
    createcalendar,

    // Hope commands
    createradio,

    // Music commands
    back,
    clear,
    filter,
    pause,
    play,
    resume,
    skip,
    loop,
    syncedlyrics,
    queue,
    nowplaying,
    remove,
    stop,
};

export const devCommands = {
    logs,
    debug,
    maintenance,
};
