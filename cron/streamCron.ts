import { CronJob } from "cron";
import { checkStreamsUpdate } from "../utils/stream/twitch";

// This cron job is used to check the stream status of users every 10 seconds.
export const streamCron = new CronJob(
    "*/10 * * * * *",
    async () => {
        await checkStreamsUpdate();
    }
);