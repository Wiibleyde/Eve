import ICAL from "ical.js"; // Using default import; types in this project don't expose members as expected

// Minimal internal shapes (fallback when library types are incomplete)
type JCal = unknown;
interface ICalTime { 
    compare(other: ICalTime): number;
    toJSDate(): Date;
}
export interface CalendarEventLike {
    startDate?: ICalTime;
    endDate?: ICalTime;
    summary?: string;
    description?: string;
    location?: string;
    uid?: string;
}
interface ICalComponentLike {
    getAllSubcomponents(name: string): unknown[];
}

// Normalized accessor to ical.js runtime API without relying on its TS types
const ICALX = ICAL as unknown as {
    parse: (data: string) => JCal;
    Component: new (jcal: JCal) => ICalComponentLike;
    Event: new (component: unknown) => CalendarEventLike;
    Time: { 
        now(): ICalTime;
        fromJSDate(date: Date, useUTC?: boolean): ICalTime;
    };
};

export class Calendar {
    private url: string;
    // Store raw jCal structure; recreate components on demand (safe with lightweight calendars)
    private jcal: JCal | null;

    constructor(url: string) {
        this.url = url;
        this.jcal = null;
    }

    static async create(url: string): Promise<Calendar> {
        const calendar = new Calendar(url);
        await calendar.refresh();
        return calendar;
    }

    public async refresh(): Promise<void> {
        const icalData = await this.fetchICalData();
        this.jcal = ICALX.parse(icalData);
    }

    private async fetchICalData(): Promise<string> {
        const response = await fetch(this.url);
        if (!response.ok) {
            throw new Error(`Failed to fetch iCal data from ${this.url}`);
        }
        return await response.text();
    }

    public getCurrentEvents(): CalendarEventLike[] {
        if (!this.jcal) {
            throw new Error("iCal data not loaded. Use Calendar.create() or refresh() first.");
        }
        const comp = new ICALX.Component(this.jcal);
        const vevents = comp.getAllSubcomponents("vevent");
        const now = ICALX.Time.now();
        const current: CalendarEventLike[] = [];
        for (const vevent of vevents) {
            const event = new ICALX.Event(vevent);
            if (event.startDate && event.endDate && event.startDate.compare(now) <= 0 && event.endDate.compare(now) >= 0) {
                current.push(event);
            }
        }
        return current;
    }

    public getUpcomingEvents(): CalendarEventLike[] {
        if (!this.jcal) {
            throw new Error("iCal data not loaded. Use Calendar.create() or refresh() first.");
        }
        const comp = new ICALX.Component(this.jcal);
        const vevents = comp.getAllSubcomponents("vevent");
        const now = ICALX.Time.now();
        const upcoming: CalendarEventLike[] = [];
        for (const vevent of vevents) {
            const event = new ICALX.Event(vevent);
            if (event.startDate && event.startDate.compare(now) > 0) {
                upcoming.push(event);
            }
        }
        // Sort ascending by start date if available
        upcoming.sort((a, b) => {
            if (!a.startDate || !b.startDate) return 0;
            return a.startDate.compare(b.startDate);
        });
        return upcoming;
    }

    public getEventsStartingSoon(minutesAhead: number): CalendarEventLike[] {
        if (!this.jcal) {
            throw new Error("iCal data not loaded. Use Calendar.create() or refresh() first.");
        }
        const comp = new ICALX.Component(this.jcal);
        const vevents = comp.getAllSubcomponents("vevent");
        const nowJs = new Date();
        const targetTime = new Date(nowJs.getTime() + minutesAhead * 60 * 1000);
        
        const soonEvents: CalendarEventLike[] = [];
        for (const vevent of vevents) {
            const event = new ICALX.Event(vevent);
            if (event.startDate) {
                const eventStartJs = event.startDate.toJSDate();
                const nowTime = nowJs.getTime();
                const eventTime = eventStartJs.getTime();
                const targetTimeMs = targetTime.getTime();
                
                // Check if event starts between now and target time
                if (eventTime > nowTime && eventTime <= targetTimeMs) {
                    soonEvents.push(event);
                }
            }
        }
        return soonEvents;
    }
}

export default Calendar;