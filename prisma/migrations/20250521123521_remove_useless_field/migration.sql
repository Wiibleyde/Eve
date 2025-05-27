/*
  Warnings:

  - You are about to drop the column `twitchChannelName` on the `Stream` table. All the data in the column will be lost.

*/
-- AlterTable
ALTER TABLE `Stream` DROP COLUMN `twitchChannelName`,
    ALTER COLUMN `twitchUserId` DROP DEFAULT;
