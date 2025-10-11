-- CreateTable
CREATE TABLE `LotoGames` (
    `uuid` VARCHAR(191) NOT NULL,
    `name` VARCHAR(191) NOT NULL,
    `isActive` BOOLEAN NOT NULL DEFAULT true,
    `ticketPrice` INTEGER NOT NULL DEFAULT 500,
    `winnerUuid` VARCHAR(191) NULL,
    `createdAt` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),

    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- CreateTable
CREATE TABLE `LotoPlayers` (
    `uuid` VARCHAR(191) NOT NULL,
    `gameUuid` VARCHAR(191) NOT NULL,
    `name` VARCHAR(191) NOT NULL,
    `lastPlay` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    INDEX `LotoPlayers_gameUuid_idx`(`gameUuid`),
    UNIQUE INDEX `LotoPlayers_gameUuid_name_key`(`gameUuid`, `name`),
    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- CreateTable
CREATE TABLE `LotoTickets` (
    `uuid` VARCHAR(191) NOT NULL,
    `playerUuid` VARCHAR(191) NOT NULL,
    `gameUuid` VARCHAR(191) NOT NULL,
    `numbers` VARCHAR(255) NOT NULL,
    `sellerId` VARCHAR(191) NOT NULL,

    INDEX `LotoTickets_playerUuid_fkey`(`playerUuid`),
    INDEX `LotoTickets_gameUuid_idx`(`gameUuid`),
    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- AddForeignKey
ALTER TABLE `LotoGames` ADD CONSTRAINT `LotoGames_winnerUuid_fkey` FOREIGN KEY (`winnerUuid`) REFERENCES `LotoPlayers`(`uuid`) ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE `LotoPlayers` ADD CONSTRAINT `LotoPlayers_gameUuid_fkey` FOREIGN KEY (`gameUuid`) REFERENCES `LotoGames`(`uuid`) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE `LotoTickets` ADD CONSTRAINT `LotoTickets_playerUuid_fkey` FOREIGN KEY (`playerUuid`) REFERENCES `LotoPlayers`(`uuid`) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE `LotoTickets` ADD CONSTRAINT `LotoTickets_gameUuid_fkey` FOREIGN KEY (`gameUuid`) REFERENCES `LotoGames`(`uuid`) ON DELETE CASCADE ON UPDATE CASCADE;
