/*
  Warnings:

  - You are about to drop the column `winnerUuid` on the `LotoGames` table. All the data in the column will be lost.

*/
-- DropForeignKey
ALTER TABLE `LotoGames` DROP FOREIGN KEY `LotoGames_winnerUuid_fkey`;

-- DropIndex
DROP INDEX `LotoGames_winnerUuid_fkey` ON `LotoGames`;

-- AlterTable
ALTER TABLE `LotoGames` DROP COLUMN `winnerUuid`;

-- CreateTable
CREATE TABLE `LotoPrizes` (
    `uuid` VARCHAR(191) NOT NULL,
    `gameUuid` VARCHAR(191) NOT NULL,
    `label` VARCHAR(255) NOT NULL,
    `position` INTEGER NOT NULL,
    `winnerPlayerUuid` VARCHAR(191) NULL,
    `winningTicketNumber` INTEGER NULL,
    `drawnAt` TIMESTAMP(6) NULL,
    `createdAt` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    `updatedAt` TIMESTAMP(6) NOT NULL,

    INDEX `LotoPrizes_gameUuid_idx`(`gameUuid`),
    INDEX `LotoPrizes_winnerPlayerUuid_idx`(`winnerPlayerUuid`),
    UNIQUE INDEX `LotoPrizes_gameUuid_position_key`(`gameUuid`, `position`),
    PRIMARY KEY (`uuid`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- AddForeignKey
ALTER TABLE `LotoPrizes` ADD CONSTRAINT `LotoPrizes_gameUuid_fkey` FOREIGN KEY (`gameUuid`) REFERENCES `LotoGames`(`uuid`) ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE `LotoPrizes` ADD CONSTRAINT `LotoPrizes_winnerPlayerUuid_fkey` FOREIGN KEY (`winnerPlayerUuid`) REFERENCES `LotoPlayers`(`uuid`) ON DELETE SET NULL ON UPDATE CASCADE;
