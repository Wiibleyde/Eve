-- CreateTable
CREATE TABLE `Streams` (
    `id` INTEGER NOT NULL AUTO_INCREMENT,
    `channelName` VARCHAR(191) NOT NULL,
    `guildId` VARCHAR(191) NOT NULL,
    `discordChannelId` VARCHAR(191) NOT NULL,
    `messageSentId` VARCHAR(191) NOT NULL,
    `roleId` VARCHAR(191) NULL,

    UNIQUE INDEX `Streams_messageSentId_key`(`messageSentId`),
    INDEX `Streams_guildId_fkey`(`guildId`),
    INDEX `Streams_channelName_fkey`(`channelName`),
    PRIMARY KEY (`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
