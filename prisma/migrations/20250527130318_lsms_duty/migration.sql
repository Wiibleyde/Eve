-- CreateTable
CREATE TABLE `LsmsDutyManager` (
    `uuid` VARCHAR(191) NOT NULL,
    `guildId` VARCHAR(191) NOT NULL,
    `channelId` VARCHAR(191) NOT NULL,
    `messageId` VARCHAR(191) NULL,
    `dutyRoleId` VARCHAR(191) NULL,
    `onCallRoleId` VARCHAR(191) NULL,
    `logsChannelId` VARCHAR(191) NULL,

    INDEX `LsmsDutyManager_guildId_idx`(`guildId`),
    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
