let maintenanceMode = false;

export function setMaintenanceMode(enabled: boolean): void {
    maintenanceMode = enabled;
}

export function isMaintenanceMode(): boolean {
    return maintenanceMode;
}

export function toggleMaintenanceMode(): boolean {
    maintenanceMode = !maintenanceMode;
    return maintenanceMode;
}
