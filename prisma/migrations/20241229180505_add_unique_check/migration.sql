/*
  Warnings:

  - A unique constraint covering the columns `[channelName,guildId]` on the table `Streams` will be added. If there are existing duplicate values, this will fail.

*/
-- CreateIndex
CREATE UNIQUE INDEX `Streams_channelName_guildId_key` ON `Streams`(`channelName`, `guildId`);
