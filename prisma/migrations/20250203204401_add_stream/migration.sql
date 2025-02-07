-- CreateTable
CREATE TABLE `Stream` (
    `uuid` VARCHAR(191) NOT NULL,
    `guildId` VARCHAR(191) NOT NULL,
    `channelId` VARCHAR(191) NOT NULL,
    `roleId` VARCHAR(191) NULL,
    `messageId` VARCHAR(191) NULL,
    `twitchChannelName` VARCHAR(191) NOT NULL,

    INDEX `Stream_guildId_idx`(`guildId`),
    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
